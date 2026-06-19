package registry

import "time"

// HealthState is the control plane's opinion about whether a backend can
// receive traffic. The data plane never decides this itself — it is TOLD.
//
//	Healthy  -> eligible to receive new requests
//	Unhealthy-> failed health checks; routing must skip it
//	Draining -> being removed gracefully; no NEW requests, existing ones finish
//
// (Draining is the "graceful shutdown" you already do when rolling a deploy:
//  stop new work, let in-flight work drain, then kill.)
type HealthState string

const (
	Healthy   HealthState = "healthy"
	Unhealthy HealthState = "unhealthy"
	Draining  HealthState = "draining"
)

// Backend == one server that can serve requests.
// Networking people also call this an "endpoint", "upstream", or "origin".
// Mentally: this is one row in your "available shards" table.
type Backend struct {
	ID      string      `json:"id"`       // stable identity, e.g. "tokyo-api-01"
	Addr    string      `json:"addr"`     // where it lives, e.g. "10.0.1.5:8080"
	DC      string      `json:"dc"`       // which datacenter, e.g. "tokyo"
	Weight  int         `json:"weight"`   // relative share of traffic (Layer 3 uses this)
	Health  HealthState `json:"health"`   // control-plane verdict; starts Healthy
	LastSeen time.Time  `json:"last_seen"`// updated by health checks (Layer 2)
}

// Cluster == a pool of interchangeable backends (replicas of ONE service).
// Also called a "backend pool" or "target group".
// Routing picks WHICH cluster; the LB algorithm picks WHICH backend inside it.
type Cluster struct {
	Name     string              `json:"name"`     // e.g. "api", "static", "checkout"
	Backends map[string]*Backend `json:"backends"` // keyed by Backend.ID
}
