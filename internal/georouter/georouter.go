package georouter

import (
	"fmt"
	"sort"
	"sync"

	"lbsim/internal/registry"
)

// Region is a named geographic origin for a request.
// In production this is derived from the client IP via a GeoIP database
// (MaxMind, ip-api, etc.). Here we simulate it as a string the caller picks.
type Region = string

// latencyTable[region][dc] = simulated one-way latency in milliseconds.
// Real systems measure this continuously with active probes (ping, HTTP HEAD)
// and update routing decisions in near-real-time. We bake in static values
// so the learning demo doesn't need network infrastructure.
var latencyTable = map[Region]map[string]int{
	"japan":     {"tokyo": 8,   "osaka": 25},
	"korea":     {"osaka": 12,  "tokyo": 30},
	"us":        {"tokyo": 105, "osaka": 110},
	"europe":    {"tokyo": 195, "osaka": 190},
	"australia": {"tokyo": 80,  "osaka": 75},
}

// AllRegions returns the regions we can simulate, sorted alphabetically.
var AllRegions = func() []string {
	rs := make([]string, 0, len(latencyTable))
	for r := range latencyTable {
		rs = append(rs, r)
	}
	sort.Strings(rs)
	return rs
}()

// DCScore is one row of the latency comparison shown to the user.
type DCScore struct {
	DC        string `json:"dc"`
	LatencyMS int    `json:"latency_ms"`
	HasHealthy bool  `json:"has_healthy"` // false = DC is down, skipped
	Chosen    bool   `json:"chosen"`
}

// Result is what POST /geo-route returns.
type Result struct {
	Backend  *registry.Backend `json:"backend"`
	Region   Region            `json:"region"`
	DCScores []DCScore         `json:"dc_scores"`
}

// GeoRouter picks the lowest-latency healthy DC for a region, then
// round-robins within that DC. Falls back to the next-best DC automatically
// if the preferred one has no healthy backends — this is the failover story
// at the global traffic management layer.
type GeoRouter struct {
	reg   *registry.Registry
	mu    sync.Mutex
	rrIdx map[string]int // "cluster/dc" → next index
}

func New(reg *registry.Registry) *GeoRouter {
	return &GeoRouter{reg: reg, rrIdx: make(map[string]int)}
}

func (g *GeoRouter) Pick(cluster, region Region) (*Result, error) {
	latencies, ok := latencyTable[region]
	if !ok {
		return nil, fmt.Errorf("unknown region %q — valid: japan, korea, us, europe, australia", region)
	}

	// Snapshot healthy backends and group them by DC.
	backends, err := g.reg.HealthyBackends(cluster)
	if err != nil {
		return nil, err
	}
	if len(backends) == 0 {
		return nil, fmt.Errorf("no healthy backends in cluster %q", cluster)
	}
	byDC := make(map[string][]*registry.Backend)
	for _, b := range backends {
		byDC[b.DC] = append(byDC[b.DC], b)
	}

	// Rank DCs by latency for this region, keeping only those with healthy backends.
	type candidate struct {
		dc      string
		latency int
	}
	var ranked []candidate
	for dc, lat := range latencies {
		if len(byDC[dc]) > 0 {
			ranked = append(ranked, candidate{dc, lat})
		}
	}
	// DCs not in the latency table (newly registered) go last.
	for dc := range byDC {
		if _, known := latencies[dc]; !known {
			ranked = append(ranked, candidate{dc, 9999})
		}
	}
	if len(ranked) == 0 {
		return nil, fmt.Errorf("no healthy DC reachable from region %q", region)
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].latency < ranked[j].latency })
	best := ranked[0]

	// Round-robin within the chosen DC.
	pool := byDC[best.dc]
	g.mu.Lock()
	key := cluster + "/" + best.dc
	i := g.rrIdx[key] % len(pool)
	g.rrIdx[key] = i + 1
	g.mu.Unlock()

	// Build the full score list (all DCs in the latency table for this region).
	scores := make([]DCScore, 0, len(latencies))
	for dc, lat := range latencies {
		scores = append(scores, DCScore{
			DC:         dc,
			LatencyMS:  lat,
			HasHealthy: len(byDC[dc]) > 0,
			Chosen:     dc == best.dc,
		})
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].LatencyMS < scores[j].LatencyMS })

	return &Result{
		Backend:  pool[i],
		Region:   region,
		DCScores: scores,
	}, nil
}
