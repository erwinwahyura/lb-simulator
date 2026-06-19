package router

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"sync"

	"lbsim/internal/registry"
)

type Algorithm string

const (
	RoundRobin     Algorithm = "round-robin"
	Weighted       Algorithm = "weighted"
	ConsistentHash Algorithm = "consistent-hash"
)

// Router sits between the data plane and the registry.
// It answers one question on every request: "which healthy backend should
// serve this?" The algorithm is swappable per call — in production you'd
// configure it per cluster (e.g. consistent-hash for session-sticky traffic,
// weighted RR for everything else).
type Router struct {
	reg   *registry.Registry
	mu    sync.Mutex
	rrIdx map[string]int // cluster -> round-robin counter
	stats map[string]int // backendID -> total dispatches
}

func New(reg *registry.Registry) *Router {
	return &Router{
		reg:   reg,
		rrIdx: make(map[string]int),
		stats: make(map[string]int),
	}
}

// Pick selects one healthy backend from the cluster.
// key is only meaningful for consistent-hash; pass "" for others.
func (r *Router) Pick(cluster string, algo Algorithm, key string) (*registry.Backend, error) {
	backends, err := r.reg.HealthyBackends(cluster)
	if err != nil {
		return nil, err
	}
	if len(backends) == 0 {
		return nil, fmt.Errorf("no healthy backends in cluster %q", cluster)
	}

	var b *registry.Backend
	switch algo {
	case Weighted:
		b = pickWeighted(backends)
	case ConsistentHash:
		b = pickConsistentHash(backends, key)
	default: // RoundRobin
		b = r.pickRoundRobin(cluster, backends)
	}

	r.mu.Lock()
	r.stats[b.ID]++
	r.mu.Unlock()

	return b, nil
}

func (r *Router) Stats() map[string]int {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]int, len(r.stats))
	for k, v := range r.stats {
		out[k] = v
	}
	return out
}

// --- Round-robin ---
// Cycles through backends in arrival order. Every backend gets exactly one
// request before any gets a second. Ignores weight.

func (r *Router) pickRoundRobin(cluster string, backends []*registry.Backend) *registry.Backend {
	r.mu.Lock()
	i := r.rrIdx[cluster] % len(backends)
	r.rrIdx[cluster] = i + 1
	r.mu.Unlock()
	return backends[i]
}

// --- Weighted random ---
// Backends with higher weight absorb proportionally more traffic.
// tokyo-api-01 (weight 5) gets ~5x more requests than osaka-api-01 (weight 1).
// The "random slot" approach: imagine a number line 0..totalWeight, pick a
// random point, walk the backends until you consume past it.

func pickWeighted(backends []*registry.Backend) *registry.Backend {
	total := 0
	for _, b := range backends {
		total += b.Weight
	}
	n := rand.Intn(total)
	for _, b := range backends {
		n -= b.Weight
		if n < 0 {
			return b
		}
	}
	return backends[len(backends)-1]
}

// --- Consistent hashing ---
// Same key → same backend (until the pool changes). Used for session affinity
// and cache locality — the same user always hits the same shard.
//
// Mental model: imagine a clock face. Each backend occupies multiple points
// (virtual nodes) around the clock. To route key K: hash K to a clock position,
// walk clockwise to the next backend point. That backend serves K.
//
// Why virtual nodes? Without them, backends cluster unevenly on the ring and
// one backend absorbs most traffic. 150 vnodes per backend gives ~even spread.
// ScyllaDB uses this exact mechanism for its token ring.

const vnodes = 150

type ringEntry struct {
	hash uint32
	id   string
}

func pickConsistentHash(backends []*registry.Backend, key string) *registry.Backend {
	// Build the ring. In production this is cached and invalidated on membership
	// change — rebuilding it on every request is O(backends * vnodes * log n).
	ring := make([]ringEntry, 0, len(backends)*vnodes)
	for _, b := range backends {
		for i := 0; i < vnodes; i++ {
			h := fnv.New32a()
			fmt.Fprintf(h, "%s#%d", b.ID, i)
			ring = append(ring, ringEntry{hash: h.Sum32(), id: b.ID})
		}
	}
	sort.Slice(ring, func(i, j int) bool { return ring[i].hash < ring[j].hash })

	// Hash the request key and find its position on the ring.
	h := fnv.New32a()
	h.Write([]byte(key))
	keyHash := h.Sum32()

	idx := sort.Search(len(ring), func(i int) bool { return ring[i].hash >= keyHash })
	if idx == len(ring) {
		idx = 0 // wrap around
	}
	target := ring[idx].id
	for _, b := range backends {
		if b.ID == target {
			return b
		}
	}
	return backends[0]
}
