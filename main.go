package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"lbsim/internal/api"
	"lbsim/internal/dispatcher"
	"lbsim/internal/georouter"
	"lbsim/internal/healthcheck"
	"lbsim/internal/l7router"
	"lbsim/internal/pipeline"
	"lbsim/internal/registry"
	"lbsim/internal/router"
)

func main() {
	reg := registry.New()

	// Seed two DCs for the "api" cluster.
	reg.CreateCluster("api")
	_ = reg.AddBackend("api", &registry.Backend{ID: "tokyo-api-01", Addr: "10.0.1.5:8080", DC: "tokyo", Weight: 5})
	_ = reg.AddBackend("api", &registry.Backend{ID: "tokyo-api-02", Addr: "10.0.1.6:8080", DC: "tokyo", Weight: 5})
	_ = reg.AddBackend("api", &registry.Backend{ID: "osaka-api-01", Addr: "10.0.2.5:8080", DC: "osaka", Weight: 3})

	// "static" cluster intentionally has no backends — used to demonstrate
	// what happens when L7 matches a rule whose cluster is empty/down.
	reg.CreateCluster("static")

	// Layer 2: health checker.
	hc := healthcheck.New(reg, 3*time.Second, 0.20, 2)
	hc.Start(context.Background())

	rtr  := router.New(reg)
	disp := dispatcher.New(rtr)
	geo  := georouter.New(reg)

	// Layer 7: seed three route rules so there is something to explore.
	//   /api/*    → api cluster   (has backends → works)
	//   /static/* → static cluster (no backends → conflict/error demo)
	//   /         → api cluster   (catch-all fallback)
	l7 := l7router.New()
	l7.AddRule(&l7router.Rule{ID: "r-api",    Method: "*",   PathPrefix: "/api/",    Cluster: "api",    Priority: 10})
	l7.AddRule(&l7router.Rule{ID: "r-static", Method: "GET", PathPrefix: "/static/", Cluster: "static", Priority: 20})
	l7.AddRule(&l7router.Rule{ID: "r-root",   Method: "*",   PathPrefix: "/",        Cluster: "api",    Priority: 99})

	pipe := pipeline.New(reg, l7, rtr)
	srv  := api.NewServer(reg, rtr, disp, geo, l7, pipe)

	addr := ":8080"
	log.Printf("lbsim %s — control plane API listening on %s", Version, addr)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}
