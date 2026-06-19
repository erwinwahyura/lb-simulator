package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"lbsim/internal/dispatcher"
	"lbsim/internal/registry"
	"lbsim/internal/router"
)

type Server struct {
	reg        *registry.Registry
	router     *router.Router
	dispatcher *dispatcher.Dispatcher
}

func NewServer(reg *registry.Registry, rtr *router.Router, d *dispatcher.Dispatcher) *Server {
	return &Server{reg: reg, router: rtr, dispatcher: d}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.dashboard)
	mux.HandleFunc("GET /clusters", s.listClusters)
	mux.HandleFunc("POST /clusters/{cluster}/backends", s.addBackend)
	mux.HandleFunc("DELETE /clusters/{cluster}/backends/{id}", s.removeBackend)
	mux.HandleFunc("PATCH /clusters/{cluster}/backends/{id}/health", s.setHealth)
	mux.HandleFunc("POST /route", s.routeRequest)
	mux.HandleFunc("GET /route/stats", s.routeStats)
	mux.HandleFunc("POST /dispatcher/start", s.dispatcherStart)
	mux.HandleFunc("POST /dispatcher/stop", s.dispatcherStop)
	mux.HandleFunc("POST /dispatcher/reset", s.dispatcherReset)
	mux.HandleFunc("GET /dispatcher/stats", s.dispatcherStats)
	return mux
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func (s *Server) listClusters(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.reg.ListClusters())
}

type addBackendReq struct {
	ID     string `json:"id"`
	Addr   string `json:"addr"`
	DC     string `json:"dc"`
	Weight int    `json:"weight"`
}

func (s *Server) addBackend(w http.ResponseWriter, r *http.Request) {
	cluster := r.PathValue("cluster")
	var req addBackendReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.ID == "" || req.Addr == "" {
		writeErr(w, http.StatusBadRequest, "id and addr are required")
		return
	}
	b := &registry.Backend{ID: req.ID, Addr: req.Addr, DC: req.DC, Weight: req.Weight}
	if err := s.reg.AddBackend(cluster, b); err != nil {
		if errors.Is(err, registry.ErrBackendExists) {
			writeErr(w, http.StatusConflict, err.Error())
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (s *Server) removeBackend(w http.ResponseWriter, r *http.Request) {
	cluster := r.PathValue("cluster")
	id := r.PathValue("id")
	if err := s.reg.RemoveBackend(cluster, id); err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type setHealthReq struct {
	Health registry.HealthState `json:"health"`
}

func (s *Server) setHealth(w http.ResponseWriter, r *http.Request) {
	cluster := r.PathValue("cluster")
	id := r.PathValue("id")
	var req setHealthReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	switch req.Health {
	case registry.Healthy, registry.Unhealthy, registry.Draining:
		// ok
	default:
		writeErr(w, http.StatusBadRequest, "health must be healthy|unhealthy|draining")
		return
	}
	if err := s.reg.SetHealth(cluster, id, req.Health); err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type routeReq struct {
	Cluster string          `json:"cluster"`
	Algo    router.Algorithm `json:"algo"`
	Key     string          `json:"key"`
}

func (s *Server) routeRequest(w http.ResponseWriter, r *http.Request) {
	var req routeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Cluster == "" {
		req.Cluster = "api"
	}
	b, err := s.router.Pick(req.Cluster, req.Algo, req.Key)
	if err != nil {
		writeErr(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"backend": b,
		"algo":    req.Algo,
		"cluster": req.Cluster,
	})
}

func (s *Server) routeStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.router.Stats())
}

type dispatchStartReq struct {
	Cluster string          `json:"cluster"`
	Algo    router.Algorithm `json:"algo"`
	RPS     int             `json:"rps"`
}

func (s *Server) dispatcherStart(w http.ResponseWriter, r *http.Request) {
	var req dispatchStartReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Cluster == "" {
		req.Cluster = "api"
	}
	if req.Algo == "" {
		req.Algo = router.RoundRobin
	}
	if !s.dispatcher.Start(req.Cluster, req.Algo, req.RPS) {
		writeErr(w, http.StatusConflict, "dispatcher already running")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (s *Server) dispatcherStop(w http.ResponseWriter, r *http.Request) {
	s.dispatcher.Stop()
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Server) dispatcherReset(w http.ResponseWriter, r *http.Request) {
	s.dispatcher.Stop()
	s.dispatcher.Reset()
	writeJSON(w, http.StatusOK, map[string]string{"status": "reset"})
}

func (s *Server) dispatcherStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.dispatcher.Stats())
}
