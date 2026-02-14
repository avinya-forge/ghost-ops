package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"ghost-ops/pkg/registry"
)

// Server handles HTTP requests.
type Server struct {
	registry *registry.Registry
	mux      *http.ServeMux
}

// NewServer creates a new API server.
func NewServer(r *registry.Registry) *Server {
	mux := http.NewServeMux()
	s := &Server{
		registry: r,
		mux:      mux,
	}

	mux.HandleFunc("/services", s.handleListServices)
	mux.HandleFunc("/reconcile", s.handleReconcile)

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

	if err := s.registry.Reconcile(r.Context()); err != nil {
		slog.Error("Failed to reconcile", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reconciliation completed"))
}
