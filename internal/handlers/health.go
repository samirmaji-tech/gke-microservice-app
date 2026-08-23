package handlers

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// Health backs the Kubernetes liveness and readiness probes.
type Health struct {
	ready atomic.Bool
}

func NewHealth() *Health { return &Health{} }

// SetReady flips readiness (false during warmup and shutdown).
func (h *Health) SetReady(v bool) { h.ready.Store(v) }

// Livez -> /healthz. "Am I alive?" Always OK unless the process is wedged.
func (h *Health) Livez(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readyz -> /readyz. "Can I serve traffic?" 503 until warmup completes.
func (h *Health) Readyz(w http.ResponseWriter, _ *http.Request) {
	if !h.ready.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not-ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
