package dispatcher

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"lbsim/internal/router"
)

// Dispatcher is the fake data plane: it fires requests at a fixed rate through
// the router and records where they land. This is what a real proxy (Envoy,
// HAProxy) would do — except real ones forward actual packets. Ours just calls
// router.Pick and counts the result.
//
// The key moment this enables: start the dispatcher, then mark a backend
// unhealthy. Watch that backend's bar stop growing while the others absorb
// its share. That redistribution is the control plane → data plane feedback
// loop made visible.
type Dispatcher struct {
	rtr *router.Router

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
	cluster string
	algo    router.Algorithm
	rps     int

	total      atomic.Int64
	bmu        sync.Mutex
	perBackend map[string]*backendCounter
}

type backendCounter struct {
	dc     string
	total  int64
	window []time.Time // timestamps of recent hits — used for rolling RPS
}

// BackendStats is the per-backend slice of the Stats response.
type BackendStats struct {
	ID    string  `json:"id"`
	DC    string  `json:"dc"`
	Total int64   `json:"total"`
	RPS   float64 `json:"rps"`
}

// Stats is what GET /dispatcher/stats returns.
type Stats struct {
	Running bool           `json:"running"`
	Total   int64          `json:"total"`
	RPS     float64        `json:"rps"`
	Algo    string         `json:"algo"`
	Cluster string         `json:"cluster"`
	RpsSet  int            `json:"rps_set"`
	Backends []BackendStats `json:"backends"`
}

func New(rtr *router.Router) *Dispatcher {
	return &Dispatcher{
		rtr:        rtr,
		perBackend: make(map[string]*backendCounter),
	}
}

// Start begins firing requests. Returns false if already running.
func (d *Dispatcher) Start(cluster string, algo router.Algorithm, rps int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.running {
		return false
	}
	if rps <= 0 {
		rps = 5
	}
	d.cluster = cluster
	d.algo = algo
	d.rps = rps
	ctx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel
	d.running = true
	go d.run(ctx)
	return true
}

func (d *Dispatcher) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.running && d.cancel != nil {
		d.cancel()
		d.running = false
	}
}

func (d *Dispatcher) Reset() {
	d.total.Store(0)
	d.bmu.Lock()
	d.perBackend = make(map[string]*backendCounter)
	d.bmu.Unlock()
}

func (d *Dispatcher) run(ctx context.Context) {
	d.mu.Lock()
	rps := d.rps
	d.mu.Unlock()

	ticker := time.NewTicker(time.Second / time.Duration(rps))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			d.fire(t)
		}
	}
}

func (d *Dispatcher) fire(t time.Time) {
	d.mu.Lock()
	cluster := d.cluster
	algo := d.algo
	d.mu.Unlock()

	// For consistent-hash: cycle through fake user keys so different users
	// land on different backends (rather than all hitting the same one).
	key := fmt.Sprintf("user-%d", rand.Intn(100))

	b, err := d.rtr.Pick(cluster, algo, key)
	if err != nil {
		return // no healthy backends — skip tick, don't panic
	}

	d.total.Add(1)

	d.bmu.Lock()
	bc, ok := d.perBackend[b.ID]
	if !ok {
		bc = &backendCounter{dc: b.DC}
		d.perBackend[b.ID] = bc
	}
	bc.total++
	bc.window = append(bc.window, t)
	// trim rolling window to last 10 seconds
	cutoff := t.Add(-10 * time.Second)
	i := 0
	for i < len(bc.window) && bc.window[i].Before(cutoff) {
		i++
	}
	bc.window = bc.window[i:]
	d.bmu.Unlock()
}

func (d *Dispatcher) Stats() Stats {
	d.mu.Lock()
	running := d.running
	algo := string(d.algo)
	cluster := d.cluster
	rps := d.rps
	d.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-10 * time.Second)

	d.bmu.Lock()
	backends := make([]BackendStats, 0, len(d.perBackend))
	windowTotal := 0
	for id, bc := range d.perBackend {
		count := 0
		for _, t := range bc.window {
			if t.After(cutoff) {
				count++
			}
		}
		windowTotal += count
		backends = append(backends, BackendStats{
			ID:    id,
			DC:    bc.dc,
			Total: bc.total,
			RPS:   float64(count) / 10.0,
		})
	}
	d.bmu.Unlock()

	return Stats{
		Running:  running,
		Total:    d.total.Load(),
		RPS:      float64(windowTotal) / 10.0,
		Algo:     algo,
		Cluster:  cluster,
		RpsSet:   rps,
		Backends: backends,
	}
}
