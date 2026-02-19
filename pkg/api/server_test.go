package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/registry"
	"ghost-ops/pkg/telemetry"
)

// MockStateStore
type MockStateStore struct {
	records map[string]protocol.ServiceRecord
}
func (m *MockStateStore) GetService(ctx context.Context, id string) (*protocol.ServiceRecord, error) {
	if r, ok := m.records[id]; ok {
		return &r, nil
	}
	return nil, nil
}
func (m *MockStateStore) UpdateService(ctx context.Context, r protocol.ServiceRecord) error {
	if m.records == nil {
		m.records = make(map[string]protocol.ServiceRecord)
	}
	m.records[r.ServiceID] = r
	return nil
}
func (m *MockStateStore) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	var list []protocol.ServiceRecord
	for _, r := range m.records {
		list = append(list, r)
	}
	return list, nil
}
func (m *MockStateStore) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}
func (m *MockStateStore) Set(ctx context.Context, key string, val []byte) error {
	return nil
}

// MockIntentSource
type MockIntentSource struct {
	blueprints []protocol.Blueprint
	index int
}
func (m *MockIntentSource) GetNextBlueprint(ctx context.Context) (*protocol.Blueprint, error) {
	if m.index >= len(m.blueprints) {
		return nil, nil
	}
	bp := m.blueprints[m.index]
	m.index++
	return &bp, nil
}

// MockEvolutionEngine
type MockEvolutionEngine struct {}
func (m *MockEvolutionEngine) Evolve(ctx context.Context, bp protocol.Blueprint) ([]byte, error) {
	return []byte("mock-wasm"), nil
}

// MockRuntimeHost
type MockRuntimeHost struct {
	modules map[string][]byte
}
func (m *MockRuntimeHost) LoadModule(ctx context.Context, id, version string, b []byte) error {
	if m.modules == nil {
		m.modules = make(map[string][]byte)
	}
	m.modules[id + "-" + version] = b
	return nil
}
func (m *MockRuntimeHost) Invoke(ctx context.Context, id, method string, p []byte) ([]byte, error) {
	return nil, nil
}
func (m *MockRuntimeHost) SetActiveVersion(ctx context.Context, id, version string) error {
	return nil
}
func (m *MockRuntimeHost) SetShadowVersion(ctx context.Context, id, version string) error {
	return nil
}
func (m *MockRuntimeHost) UnsetShadowVersion(ctx context.Context, id string) error {
	return nil
}
func (m *MockRuntimeHost) UnloadVersion(ctx context.Context, id, version string) error {
	if m.modules != nil {
		delete(m.modules, id + "-" + version)
	}
	return nil
}
func (m *MockRuntimeHost) Close(ctx context.Context) error {
	return nil
}

func TestServer_Services(t *testing.T) {
	store := &MockStateStore{records: map[string]protocol.ServiceRecord{
		"svc-1": {ServiceID: "svc-1"},
	}}
	collector := telemetry.NewInMemoryCollector()
	reg := registry.NewRegistry(store, &MockEvolutionEngine{}, &MockIntentSource{}, &MockRuntimeHost{}, collector)
	server := NewServer(reg, collector)

	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	// Check JSON body if needed
}

func TestServer_Reconcile(t *testing.T) {
	store := &MockStateStore{records: make(map[string]protocol.ServiceRecord)}
	collector := telemetry.NewInMemoryCollector()
	reg := registry.NewRegistry(store, &MockEvolutionEngine{}, &MockIntentSource{}, &MockRuntimeHost{modules: make(map[string][]byte)}, collector)
	server := NewServer(reg, collector)

	req := httptest.NewRequest(http.MethodPost, "/reconcile", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestServer_Metrics(t *testing.T) {
	store := &MockStateStore{records: make(map[string]protocol.ServiceRecord)}
	collector := telemetry.NewInMemoryCollector()
	// Increment a counter
	collector.Counter("test_counter", 1, nil)

	reg := registry.NewRegistry(store, &MockEvolutionEngine{}, &MockIntentSource{}, &MockRuntimeHost{}, collector)
	server := NewServer(reg, collector)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected body, got empty")
	}
}
