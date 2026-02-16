package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/registry"
)

// Server handles HTTP requests.
type Server struct {
	registry  *registry.Registry
	collector protocol.MetricsCollector
	mux       *http.ServeMux
}

// NewServer creates a new API server.
func NewServer(r *registry.Registry, c protocol.MetricsCollector) *Server {
	mux := http.NewServeMux()
	s := &Server{
		registry:  r,
		collector: c,
		mux:       mux,
	}

	mux.HandleFunc("/services", s.handleListServices)
	mux.HandleFunc("/reconcile", s.handleReconcile)
	mux.HandleFunc("/metrics", s.handleMetrics)

	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleListServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	services, err := s.registry.ListServices(r.Context())
	if err != nil {
		slog.Error("Failed to list services", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(services); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}

func (s *Server) handleReconcile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	for {
		processed, err := s.registry.Reconcile(r.Context())
		if err != nil {
			slog.Error("Failed to reconcile", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !processed {
			break
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reconciliation completed"))
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type snapshotter interface {
		Snapshot() map[string]interface{}
	}

	var metrics map[string]interface{}
	if s, ok := s.collector.(snapshotter); ok {
		metrics = s.Snapshot()
	} else {
		metrics = map[string]interface{}{"error": "collector does not support snapshot"}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		slog.Error("Failed to encode metrics", "error", err)
	}
}
