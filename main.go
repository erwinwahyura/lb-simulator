package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"lbsim/internal/api"
	"lbsim/internal/healthcheck"
	"lbsim/internal/registry"
	"lbsim/internal/router"
)

func main() {
	reg := registry.New()

	// Seed a little world so there's something to query immediately.
	// Two datacenters, one "api" cluster with replicas in each.
	reg.CreateCluster("api")
	_ = reg.AddBackend("api", &registry.Backend{ID: "tokyo-api-01", Addr: "10.0.1.5:8080", DC: "tokyo", Weight: 5})
	_ = reg.AddBackend("api", &registry.Backend{ID: "tokyo-api-02", Addr: "10.0.1.6:8080", DC: "tokyo", Weight: 5})
	_ = reg.AddBackend("api", &registry.Backend{ID: "osaka-api-01", Addr: "10.0.2.5:8080", DC: "osaka", Weight: 3})

	// Layer 2: health checker.
	// Probes every 3 s. Each probe has a 20% chance of "failing".
	// A backend needs 2 consecutive failures before it's marked unhealthy
	// (flap protection — one bad probe doesn't immediately pull it from rotation).
	hc := healthcheck.New(reg, 3*time.Second, 0.20, 2)
	hc.Start(context.Background())

	rtr := router.New(reg)
	srv := api.NewServer(reg, rtr)

	addr := ":8080"
	log.Printf("lbsim %s — control plane API listening on %s", Version, addr)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}
