package registry

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrClusterNotFound = errors.New("cluster not found")
	ErrBackendNotFound = errors.New("backend not found")
	ErrBackendExists   = errors.New("backend already exists")
)

// Registry is the control plane's source of truth: every cluster and backend
// the system knows about. EVERYTHING later (health checker, router, failover)
// reads from here.
//
// Why a mutex: the HTTP API mutates this from one goroutine, the health checker
// mutates it from another, and the router READS it from request-handling
// goroutines — all concurrently. This is the exact connection-pool / shared-map
// concurrency you handle in Go services already; nothing networking-specific.
//
// Why in-memory (not Redis/Scylla yet): Layer 1 keeps it boring on purpose so
// the routing concepts stay in focus. Swapping this for Redis later is a drop-in
// — the method set is the seam. Swap the implementation for Redis/Scylla and
// nothing above it changes.
type Registry struct {
	mu       sync.RWMutex
	clusters map[string]*Cluster
}

func New() *Registry {
	return &Registry{clusters: make(map[string]*Cluster)}
}

// --- Cluster operations ---

func (r *Registry) CreateCluster(name string) *Cluster {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.clusters[name]; ok {
		return c // idempotent: creating an existing cluster is a no-op
	}
	c := &Cluster{Name: name, Backends: make(map[string]*Backend)}
	r.clusters[name] = c
	return c
}

func (r *Registry) ListClusters() []*Cluster {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Cluster, 0, len(r.clusters))
	for _, c := range r.clusters {
		out = append(out, c)
	}
	return out
}

// ClusterNames returns a snapshot of cluster name strings — safe to iterate
// outside the lock because it's a plain []string copy.
func (r *Registry) ClusterNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.clusters))
	for name := range r.clusters {
		names = append(names, name)
	}
	return names
}

// ListBackends returns a snapshot of all backends in a cluster as value copies.
// The health checker reads these without holding the registry lock, so we copy
// the structs here to avoid a data race on b.Health.
func (r *Registry) ListBackends(cluster string) ([]*Backend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.clusters[cluster]
	if !ok {
		return nil, ErrClusterNotFound
	}
	out := make([]*Backend, 0, len(c.Backends))
	for _, b := range c.Backends {
		cp := *b
		out = append(out, &cp)
	}
	return out, nil
}

// --- Backend operations ---

// AddBackend registers a new server into a cluster (creating the cluster if
// needed). A freshly-registered backend starts Healthy — in a real system you
// might start it Unhealthy until the first health check passes, which is a
// safer default. We'll revisit that nuance in Layer 2.
func (r *Registry) AddBackend(cluster string, b *Backend) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.clusters[cluster]
	if !ok {
		c = &Cluster{Name: cluster, Backends: make(map[string]*Backend)}
		r.clusters[cluster] = c
	}
	if _, exists := c.Backends[b.ID]; exists {
		return ErrBackendExists
	}
	if b.Weight <= 0 {
		b.Weight = 1 // sane default so weighted routing never divides by zero
	}
	b.Health = Healthy
	b.LastSeen = time.Now()
	c.Backends[b.ID] = b
	return nil
}

// RemoveBackend hard-removes a backend (deregister). Compare with SetHealth
// (Draining) for the graceful path.
func (r *Registry) RemoveBackend(cluster, backendID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.clusters[cluster]
	if !ok {
		return ErrClusterNotFound
	}
	if _, ok := c.Backends[backendID]; !ok {
		return ErrBackendNotFound
	}
	delete(c.Backends, backendID)
	return nil
}

// SetHealth lets the health checker (or an operator) flip a backend's state.
// This is the lever failover pulls: mark Unhealthy and the router stops
// selecting it WITHOUT removing it, so it can come back.
func (r *Registry) SetHealth(cluster, backendID string, h HealthState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.clusters[cluster]
	if !ok {
		return ErrClusterNotFound
	}
	b, ok := c.Backends[backendID]
	if !ok {
		return ErrBackendNotFound
	}
	b.Health = h
	b.LastSeen = time.Now()
	return nil
}

// SetDCHealth flips every backend in a datacenter to the given health state.
// This is the "kill a DC" lever: one call takes down (or recovers) an entire
// site. Returns how many backends were affected.
func (r *Registry) SetDCHealth(cluster, dc string, h HealthState) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.clusters[cluster]
	if !ok {
		return 0, ErrClusterNotFound
	}
	n := 0
	for _, b := range c.Backends {
		if b.DC == dc {
			b.Health = h
			b.LastSeen = time.Now()
			n++
		}
	}
	return n, nil
}

// HealthyBackends returns only the backends a router may select from.
// This is THE read path the routing layer will call on every request, so it
// returns a copied slice (snapshot) — the caller iterates without holding the
// lock, and concurrent mutations can't corrupt its loop.
func (r *Registry) HealthyBackends(cluster string) ([]*Backend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.clusters[cluster]
	if !ok {
		return nil, ErrClusterNotFound
	}
	out := make([]*Backend, 0, len(c.Backends))
	for _, b := range c.Backends {
		if b.Health == Healthy {
			out = append(out, b)
		}
	}
	return out, nil
}
