package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ghost-ops/pkg/config"
	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/registry"
	"ghost-ops/pkg/telemetry"
)

func TestClusterHealth(t *testing.T) {
	ctx := context.Background()
	stateStore := protocol.NewInMemoryStateStore()
	err := stateStore.UpdateService(ctx, protocol.ServiceRecord{
		ServiceID:     "test-svc",
		ActiveVersion: 1,
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	engine := evolution.NewMockEvolutionEngine("")
	source := &MockIntentSource{}
	collector := telemetry.NewInMemoryCollector()
	host := &MockRuntimeHost{modules: make(map[string][]byte)}
	cfg := config.RegistryConfig{MaxActiveServices: 10}

	reg := registry.NewRegistry(stateStore, engine, source, host, collector, nil, cfg)

	srv := NewServer(reg, collector, 1024)

	req, _ := http.NewRequest("GET", "/cluster/health", nil)
	rr := httptest.NewRecorder()

	srv.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var health map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &health)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if health["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %v", health["status"])
	}

	total := int(health["total_active_services"].(float64))
	if total != 1 {
		t.Errorf("expected 1 active service, got %v", total)
	}
}
