# lbsim — Load Balancer Control Plane Simulator

A learn-by-building control plane for a software load balancer, in Go.
Goal: internalize traffic routing, health checking, failover, and global
traffic management — without the kernel-networking rabbit hole.

We build the **control plane** (the brain) and a **fake data plane** (a
dispatcher that asks "which backend?"). All the interesting logic lives in
the control plane.

## Vocabulary map (networking term -> thing you already know)

| Networking term        | What it really is                                  |
|------------------------|----------------------------------------------------|
| Data plane             | The fast path that forwards packets (Envoy/HAProxy). Here: a Go dispatcher. |
| Control plane          | The brain: state + decisions + config push. Your Go API. |
| Backend / endpoint / upstream / origin | One server. One row in your "available shards" table. |
| Cluster / target group / backend pool  | A pool of interchangeable replicas of one service. |
| Listener               | The front door IP:port traffic arrives on.         |
| Health checking        | A ticker goroutine probing backends, flipping a bool. |
| Draining               | Graceful removal: no new requests, finish in-flight. |
| Consistent hashing     | ScyllaDB's token ring, applied to picking a backend per user. |
| xDS                    | Config push from brain to fast path. Like your precompute->Redis layer. |
| L4 routing             | Decide using IP/port only (TCP/UDP). Fast, dumb.    |
| L7 routing             | Decide using HTTP path/headers. Smart, costs parsing.|

## Run it

```bash
go run .
```

Open **http://localhost:8080** for the live dashboard (auto-refreshes every 2 s).

In another terminal (raw API):

```bash
# list clusters (seeded with an "api" cluster across tokyo + osaka)
curl -s localhost:8080/clusters | jq

# register a new backend
curl -s -X POST localhost:8080/clusters/api/backends \
  -d '{"id":"osaka-api-02","addr":"10.0.2.6:8080","dc":"osaka","weight":3}' | jq

# mark one unhealthy
curl -s -X PATCH localhost:8080/clusters/api/backends/tokyo-api-01/health \
  -d '{"health":"unhealthy"}' | jq

# route a request (round-robin | weighted | consistent-hash)
curl -s -X POST localhost:8080/route \
  -d '{"cluster":"api","algo":"consistent-hash","key":"user-42"}' | jq

# dispatch counts per backend
curl -s localhost:8080/route/stats | jq

# deregister
curl -s -X DELETE localhost:8080/clusters/api/backends/osaka-api-02
```

## Versioning

Every push bumps the version in `version.go`. Version log:

| Version | Layer | What shipped |
|---------|-------|-------------|
| v0.1.0  | L1    | Registry + Web API (clusters, backends, CRUD) |
| v0.2.0  | L2    | Health checker: ticker goroutines, flap protection (2 consecutive fails to flip) |
| v0.3.0  | L3    | Routing algorithms: round-robin, weighted, consistent-hash; dashboard routing panel + dispatch stats |
| v0.4.0  | L4    | Live dispatcher: continuous request stream, per-backend traffic bars, RPS counter, start/stop/reset |
| v0.4.1  | —     | Hover tooltips on every interactive element — algorithms, health states, buttons, labels, traffic bars |
| v0.5.0  | L5    | Failover & draining: SetDCHealth bulk API, Failover Lab panel, per-DC cards, event log |

## Layer roadmap

- **Layer 0** — foundations: data plane vs control plane. ✅ (conceptual)
- **Layer 1** — registry + Web API. The control plane's state store. ✅ `v0.1.0`
- **Layer 2** — health checking: ticker goroutines, up/down transitions, flap protection. ✅ `v0.2.0`
- **Layer 3** — routing algorithms: round-robin, weighted, consistent hashing. ✅ `v0.3.0`
- **Layer 4** — the dispatcher (fake data plane) + request distribution metrics. ✅ `v0.4.0`
- **Layer 5** — failover & draining: kill a backend/DC, watch traffic rebalance live. ✅ `v0.5.0`
- **Layer 6** — global traffic management: GeoDNS / latency-based DC selection.

## Architecture note

The registry's method set is the seam. Today it's an in-memory map behind a
RWMutex. Swap that implementation for Redis/Scylla and nothing above it changes
— the control plane's state backing store is pluggable.
