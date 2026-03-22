package gateway

import (
	"context"
	"io"
	"net/http"
	"strings"

	"ghost-ops/pkg/protocol"
)

// Gateway handles Layer 7 routing to WASM services based on paths and headers.
type Gateway struct {
	runtimeHost protocol.RuntimeHost
}

// NewGateway creates a new Gateway instance.
func NewGateway(host protocol.RuntimeHost) *Gateway {
	return &Gateway{
		runtimeHost: host,
	}
}

// ServeHTTP implements http.Handler.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Determine the target serviceID based on path and headers
	serviceID := g.routeRequest(r)

	if serviceID == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// For Gateway logic we just pass the request body to the WASM invocation.
	// We'll assume the `method` parameter is the HTTP method (e.g., "GET", "POST").
	method := r.Method

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Payload Too Large", http.StatusRequestEntityTooLarge)
		return
	}
	defer r.Body.Close()

	// Invoke the WASM service
	res, err := g.runtimeHost.Invoke(context.Background(), serviceID, method, body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func (g *Gateway) routeRequest(r *http.Request) string {
	path := r.URL.Path

	// 1. Header-Based Routing (Task 901.3)
	// Example: X-Beta: true -> beta-service
	if r.Header.Get("X-Beta") == "true" {
		return "beta-service"
	}

	// 2. Path-Based Routing (Task 901.2)
	// Example: /api/v1/auth -> auth-service
	if strings.HasPrefix(path, "/api/v1/") {
		parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0] + "-service"
		}
	}

	return ""
}
