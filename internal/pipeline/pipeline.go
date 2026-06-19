package pipeline

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"sync"
	"time"

	"lbsim/internal/l7router"
	"lbsim/internal/registry"
	"lbsim/internal/router"
)

// Trace is the full record of one request's journey through the LB stack.
// This is what a distributed tracing system (Jaeger, Zipkin) stores and
// what you'd query when debugging "why did this request fail?"
type Trace struct {
	RequestID  string            `json:"request_id"`
	Timestamp  string            `json:"timestamp"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	ClientIP   string            `json:"client_ip"`
	L4Cluster  string            `json:"l4_cluster"`  // what L4 (IP hash) would pick
	L7Cluster  string            `json:"l7_cluster"`  // what L7 (path match) picked
	MatchedRule string           `json:"matched_rule"` // which rule fired
	Conflict   bool              `json:"conflict"`    // L4 and L7 disagree
	Backend    *registry.Backend `json:"backend,omitempty"`
	LatencyMS  int               `json:"latency_ms"`
	Headers    map[string]string `json:"headers"` // headers injected by the LB
	Error      string            `json:"error,omitempty"`
}

// Pipeline models the full request path:
//
//	client → [L4 pre-routing by IP] → [L7 path matching] → [conflict check] → backend
//
// The OSI clash happens when L4 and L7 pick different clusters.
// In a real LB chain (e.g. AWS NLB → ALB), the L4 layer can't see HTTP
// paths so it load-balances blindly. If L7 has a rule that contradicts what
// L4 sent, the request ends up at the wrong cluster (or fails).
type Pipeline struct {
	reg *registry.Registry
	l7  *l7router.L7Router
	rtr *router.Router

	mu        sync.Mutex
	traces    []*Trace
	maxTraces int
}

func New(reg *registry.Registry, l7 *l7router.L7Router, rtr *router.Router) *Pipeline {
	return &Pipeline{reg: reg, l7: l7, rtr: rtr, maxTraces: 50}
}

// Process runs one simulated request through the full pipeline and records the trace.
func (p *Pipeline) Process(method, path, clientIP string) *Trace {
	reqID := fmt.Sprintf("req-%06x", rand.Intn(0xffffff))

	trace := &Trace{
		RequestID: reqID,
		Timestamp: time.Now().Format("15:04:05.000"),
		Method:    method,
		Path:      path,
		ClientIP:  clientIP,
		Headers:   make(map[string]string),
	}

	// Standard headers every LB injects — these travel with the request to
	// the backend and appear in its access logs.
	trace.Headers["X-Request-ID"]    = reqID
	trace.Headers["X-Forwarded-For"] = clientIP
	trace.Headers["Via"]             = "1.1 lbsim-edge (lbsim)"

	// ── Step 1: L4 pre-routing ───────────────────────────────────────────
	// L4 only sees IP:port. It consistent-hashes the client IP across ALL
	// healthy backends from ALL clusters — it has no idea what /api/ means.
	l4Cluster := p.l4Cluster(clientIP)
	trace.L4Cluster = l4Cluster
	trace.Headers["X-L4-Cluster"] = l4Cluster

	// ── Step 2: L7 path matching ─────────────────────────────────────────
	// L7 reads the HTTP path and matches it against route rules.
	rule := p.l7.Match(method, path)
	if rule != nil {
		trace.L7Cluster = rule.Cluster
		trace.MatchedRule = fmt.Sprintf("%s  %s %s → %s", rule.ID, rule.Method, rule.PathPrefix, rule.Cluster)
	} else {
		trace.L7Cluster = l4Cluster // no L7 rule → fall through to whatever L4 picked
		trace.MatchedRule = "none (L4 passthrough)"
	}
	trace.Headers["X-L7-Cluster"]    = trace.L7Cluster
	trace.Headers["X-Matched-Rule"]  = trace.MatchedRule

	// ── Step 3: Conflict detection ───────────────────────────────────────
	// Conflict = L4 and L7 picked different clusters.
	// This happens in real systems when, e.g.:
	//   - L4 sticky sessions hash user X to the "api" cluster
	//   - but the path /static/ should go to the "static" cluster
	// L7 wins (it knows more), but if the L7-matched cluster has no healthy
	// backends, the request fails — and L4's choice can't save it.
	trace.Conflict = l4Cluster != "" && rule != nil && l4Cluster != trace.L7Cluster
	if trace.Conflict {
		trace.Headers["X-Conflict"] = fmt.Sprintf("L4→%s vs L7→%s (L7 wins)", l4Cluster, trace.L7Cluster)
	}

	// ── Step 4: Backend selection ────────────────────────────────────────
	// L7 decision wins. Pick a backend from the L7-matched cluster.
	target := trace.L7Cluster
	b, err := p.rtr.Pick(target, router.RoundRobin, "")
	if err != nil {
		trace.Error = err.Error()
		trace.Headers["X-Error"] = err.Error()
	} else {
		trace.Backend = b
		trace.LatencyMS = rand.Intn(13) + 2 // simulate 2–14 ms backend latency
		trace.Headers["X-Upstream"]         = b.ID
		trace.Headers["X-Upstream-DC"]      = b.DC
		trace.Headers["X-Upstream-Latency"] = fmt.Sprintf("%dms", trace.LatencyMS)
	}

	p.mu.Lock()
	p.traces = append([]*Trace{trace}, p.traces...)
	if len(p.traces) > p.maxTraces {
		p.traces = p.traces[:p.maxTraces]
	}
	p.mu.Unlock()

	return trace
}

func (p *Pipeline) Traces() []*Trace {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]*Trace, len(p.traces))
	copy(out, p.traces)
	return out
}

// l4Cluster simulates what an L4 LB does: consistent-hash on client IP
// across ALL backends in ALL clusters, then return which cluster that backend
// belongs to. It has no path information — pure IP:port decision.
func (p *Pipeline) l4Cluster(clientIP string) string {
	type entry struct {
		cluster string
		id      string
	}
	var all []entry
	for _, name := range p.reg.ClusterNames() {
		backends, _ := p.reg.HealthyBackends(name)
		for _, b := range backends {
			all = append(all, entry{name, b.ID})
		}
	}
	if len(all) == 0 {
		return ""
	}
	h := fnv32(clientIP)
	return all[h%uint32(len(all))].cluster
}

func fnv32(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
