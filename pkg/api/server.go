package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ghost-ops/pkg/registry"
)

// APIServer provides an HTTP API for the GhostOps system.
type APIServer struct {
	registry *registry.ServiceRegistry
	mux      *http.ServeMux
}

// NewAPIServer creates a new APIServer.
func NewAPIServer(r *registry.ServiceRegistry) *APIServer {
	s := &APIServer{
		registry: r,
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *APIServer) registerRoutes() {
	s.mux.HandleFunc("/services", s.handleListServices)
	s.mux.HandleFunc("/reconcile", s.handleReconcile)
}

// Serve starts the HTTP server.
func (s *APIServer) Serve(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

// ServeHTTP implements http.Handler for testing.
func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *APIServer) handleListServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	services, err := s.registry.ListServices(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list services: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(services); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

func (s *APIServer) handleReconcile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	processed, err := s.registry.Reconcile(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Reconciliation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if processed {
		w.Write([]byte(`{"status": "ok", "message": "reconciliation complete"}`))
	} else {
		w.Write([]byte(`{"status": "ok", "message": "no new blueprints"}`))
	}
}
