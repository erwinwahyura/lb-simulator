package healthcheck

import (
	"context"
	"log"
	"math/rand"
	"time"

	"lbsim/internal/registry"
)

// Checker runs a simulated health probe against every backend on a ticker.
//
// Real systems do an actual TCP dial or HTTP GET here. We simulate: each probe
// independently rolls a random float against failProb. The key insight is the
// STATE MACHINE around the result — a single blip doesn't flip a backend down
// (flap protection), and a single recovery immediately flips it back up.
//
// Flap protection matters in production: without it, a backend that's 10% slow
// will flicker in/out of rotation every few seconds, making things worse.
type Checker struct {
	reg       *registry.Registry
	interval  time.Duration
	failProb  float64 // 0.0–1.0: probability a single probe "fails"
	threshold int     // consecutive failures required before marking unhealthy
}

func New(reg *registry.Registry, interval time.Duration, failProb float64, threshold int) *Checker {
	return &Checker{reg: reg, interval: interval, failProb: failProb, threshold: threshold}
}

func (c *Checker) Start(ctx context.Context) {
	go c.run(ctx)
}

// backendState is the health checker's own opinion — separate from the registry.
// We track it here so we never need to read b.Health from the registry
// (which would race with concurrent writers).
type backendState struct {
	consecutiveFailures int
	markedDown          bool
}

func (c *Checker) run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// keyed by "clusterName/backendID"
	states := make(map[string]*backendState)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.tick(states)
		}
	}
}

func (c *Checker) tick(states map[string]*backendState) {
	for _, clusterName := range c.reg.ClusterNames() {
		backends, err := c.reg.ListBackends(clusterName)
		if err != nil {
			continue
		}
		for _, b := range backends {
			key := clusterName + "/" + b.ID
			st, ok := states[key]
			if !ok {
				st = &backendState{}
				states[key] = st
			}

			probeOK := rand.Float64() >= c.failProb

			if probeOK {
				// probe passed: reset failure streak
				if st.markedDown {
					st.markedDown = false
					_ = c.reg.SetHealth(clusterName, b.ID, registry.Healthy)
					log.Printf("[hc] %s → healthy (recovered)", b.ID)
				}
				st.consecutiveFailures = 0
			} else {
				// probe failed: increment streak
				st.consecutiveFailures++
				log.Printf("[hc] %s probe failed (%d/%d)", b.ID, st.consecutiveFailures, c.threshold)
				if !st.markedDown && st.consecutiveFailures >= c.threshold {
					st.markedDown = true
					_ = c.reg.SetHealth(clusterName, b.ID, registry.Unhealthy)
					log.Printf("[hc] %s → unhealthy (flap protection tripped)", b.ID)
				}
			}
		}
	}
}
