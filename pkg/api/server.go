package api

import (
	"encoding/json"
	"io"
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
	mux.HandleFunc("GET /services/{id}/logs", validateServiceIDMiddleware(s.handleServiceLogs))
	mux.HandleFunc("POST /services/{id}/invoke", validateServiceIDMiddleware(s.handleServiceInvoke))
	mux.HandleFunc("/reconcile", s.handleReconcile)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/healthz", s.handleHealthz)

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
		s.writeError(w, err)
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
			s.writeError(w, err)
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

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleServiceLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing service ID", http.StatusBadRequest)
		return
	}

	logs, err := s.registry.GetLogs(r.Context(), id)
	if err != nil {
		s.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(logs)
}

func (s *Server) handleServiceInvoke(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing service ID", http.StatusBadRequest)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, protocol.NewAppError(protocol.ErrInvalidInput.Code(), "failed to read payload", err))
		return
	}

	method := r.URL.Query().Get("method")
	if method == "" {
		method = "default"
	}

	res, err := s.registry.Invoke(r.Context(), id, method, payload)
	if err != nil {
		s.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(res)
}

func (s *Server) writeError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	message := "Internal server error"

	if appErr, ok := err.(protocol.AppError); ok {
		switch appErr.Code() {
		case protocol.ErrNotFound.Code():
			statusCode = http.StatusNotFound
		case protocol.ErrInvalidInput.Code():
			statusCode = http.StatusBadRequest
		case protocol.ErrTimeout.Code():
			statusCode = http.StatusGatewayTimeout
		case protocol.ErrConflict.Code():
			statusCode = http.StatusConflict
		case protocol.ErrUnauthorized.Code():
			statusCode = http.StatusUnauthorized
		}
		// Use the full error message for context
		message = appErr.Error()
	}

	slog.Error("Request failed", "error", err, "status_code", statusCode)

	// Log stack trace for internal errors
	if statusCode == http.StatusInternalServerError {
		if appErr, ok := err.(protocol.AppError); ok {
			stack := appErr.StackTrace()
			if stack != "" {
				slog.Error("Internal error stack trace", "stack", stack)
			}
		}
	}

	http.Error(w, message, statusCode)
}
