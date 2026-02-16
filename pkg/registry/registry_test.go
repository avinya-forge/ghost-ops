package registry

import (
	"context"
	"testing"

	"ghost-ops/pkg/protocol"
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
func (m *MockRuntimeHost) LoadModule(ctx context.Context, id string, b []byte) error {
	m.modules[id] = b
	return nil
}
func (m *MockRuntimeHost) Invoke(ctx context.Context, id, method string, p []byte) ([]byte, error) {
	return nil, nil
}
func (m *MockRuntimeHost) UnloadModule(ctx context.Context, id string) error {
	delete(m.modules, id)
	return nil
}


func TestRegistry_Reconcile(t *testing.T) {
	store := &MockStateStore{records: make(map[string]protocol.ServiceRecord)}
	engine := &MockEvolutionEngine{}
	source := &MockIntentSource{
		blueprints: []protocol.Blueprint{
			{ServiceID: "svc-1", Intent: "do something"},
		},
	}
	runtime := &MockRuntimeHost{modules: make(map[string][]byte)}
	collector := telemetry.NewInMemoryCollector()

	reg := NewRegistry(store, engine, source, runtime, collector)
	ctx := context.Background()

	// First call should process one blueprint
	processed, err := reg.Reconcile(ctx)
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if !processed {
		t.Fatal("Expected Reconcile to return true for processed blueprint")
	}

	// Verify Store
	rec, _ := store.GetService(ctx, "svc-1")
	if rec == nil {
		t.Fatal("Service not found in store")
	}
	if rec.CurrentState != protocol.StateActive {
		t.Errorf("Expected Active, got %s", rec.CurrentState)
	}

	// Verify Runtime
	if _, ok := runtime.modules["svc-1"]; !ok {
		t.Fatal("Service not found in runtime")
	}

	// Second call should find no more blueprints
	processed, err = reg.Reconcile(ctx)
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if processed {
		t.Fatal("Expected Reconcile to return false for no more blueprints")
	}
}
