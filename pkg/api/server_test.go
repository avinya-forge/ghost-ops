package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/registry"
)

// Minimal mocks for dependencies of ServiceRegistry

type MockIntentSource struct {
	Blueprints []protocol.Blueprint
	index      int
}

func (m *MockIntentSource) GetNextBlueprint(ctx context.Context) (*protocol.Blueprint, error) {
	if m.index >= len(m.Blueprints) {
		return nil, nil
	}
	bp := m.Blueprints[m.index]
	m.index++
	return &bp, nil
}

type MockEvolutionEngine struct{}

func (m *MockEvolutionEngine) Evolve(ctx context.Context, blueprint protocol.Blueprint) ([]byte, error) {
	return []byte("mock-wasm"), nil
}

type MockRuntimeHost struct{}

func (m *MockRuntimeHost) LoadModule(ctx context.Context, serviceID string, wasm []byte) error { return nil }
func (m *MockRuntimeHost) Invoke(ctx context.Context, serviceID, method string, payload []byte) ([]byte, error) { return nil, nil }
func (m *MockRuntimeHost) UnloadModule(ctx context.Context, serviceID string) error { return nil }

type MockStateStore struct {
	services []protocol.ServiceRecord
}

func (m *MockStateStore) GetService(ctx context.Context, serviceID string) (*protocol.ServiceRecord, error) { return nil, nil }
func (m *MockStateStore) UpdateService(ctx context.Context, record protocol.ServiceRecord) error {
	m.services = append(m.services, record)
	return nil
}
func (m *MockStateStore) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	return m.services, nil
}

func TestAPIServer_ListServices(t *testing.T) {
	store := &MockStateStore{
		services: []protocol.ServiceRecord{
			{ServiceID: "svc-1", CurrentState: protocol.StateActive},
		},
	}
	reg := registry.NewServiceRegistry(
		&MockIntentSource{},
		&MockEvolutionEngine{},
		&MockRuntimeHost{},
		store,
	)
	server := NewAPIServer(reg)

	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var services []protocol.ServiceRecord
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}
	if services[0].ServiceID != "svc-1" {
		t.Errorf("Expected svc-1, got %s", services[0].ServiceID)
	}
}

func TestAPIServer_Reconcile(t *testing.T) {
	intent := &MockIntentSource{
		Blueprints: []protocol.Blueprint{
			{ServiceID: "svc-new", Intent: "intent"},
		},
	}
	store := &MockStateStore{}
	reg := registry.NewServiceRegistry(
		intent,
		&MockEvolutionEngine{},
		&MockRuntimeHost{},
		store,
	)
	server := NewAPIServer(reg)

	req := httptest.NewRequest(http.MethodPost, "/reconcile", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify service was added to store via reconciliation
	if len(store.services) != 1 {
		t.Errorf("Expected 1 service in store, got %d", len(store.services))
	}
}
