package registry

import (
	"context"
	"fmt"
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
	modules        map[string][]byte
	activeVersions map[string]string
}
func (m *MockRuntimeHost) LoadModule(ctx context.Context, id, version string, b []byte) error {
	if m.modules == nil {
		m.modules = make(map[string][]byte)
	}
	uniqueName := fmt.Sprintf("%s-%s", id, version)
	m.modules[uniqueName] = b
	return nil
}
func (m *MockRuntimeHost) Invoke(ctx context.Context, id, method string, p []byte) ([]byte, error) {
	return nil, nil
}
func (m *MockRuntimeHost) SetActiveVersion(ctx context.Context, id, version string) error {
	if m.activeVersions == nil {
		m.activeVersions = make(map[string]string)
	}
	m.activeVersions[id] = version
	return nil
}
func (m *MockRuntimeHost) UnloadVersion(ctx context.Context, id, version string) error {
	uniqueName := fmt.Sprintf("%s-%s", id, version)
	delete(m.modules, uniqueName)
	return nil
}
func (m *MockRuntimeHost) Close(ctx context.Context) error {
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
	runtime := &MockRuntimeHost{
		modules: make(map[string][]byte),
		activeVersions: make(map[string]string),
	}
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

	// Verify Runtime - Version 1 should be loaded
	if _, ok := runtime.modules["svc-1-1"]; !ok {
		t.Errorf("Service version 1 not found in runtime modules: %v", runtime.modules)
	}
	// Verify Active Version
	if ver, ok := runtime.activeVersions["svc-1"]; !ok || ver != "1" {
		t.Errorf("Service active version mismatch: got %v, want 1", ver)
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

func TestRegistry_Reconcile_Versioning(t *testing.T) {
	store := &MockStateStore{records: make(map[string]protocol.ServiceRecord)}
	engine := &MockEvolutionEngine{}
	source := &MockIntentSource{
		blueprints: []protocol.Blueprint{
			{ServiceID: "svc-1", Intent: "v1"},
			{ServiceID: "svc-1", Intent: "v2"},
		},
	}
	runtime := &MockRuntimeHost{
		modules: make(map[string][]byte),
		activeVersions: make(map[string]string),
	}
	collector := telemetry.NewInMemoryCollector()

	reg := NewRegistry(store, engine, source, runtime, collector)
	ctx := context.Background()

	// First call - Version 1
	processed, err := reg.Reconcile(ctx)
	if err != nil {
		t.Fatalf("Reconcile 1 failed: %v", err)
	}
	if !processed {
		t.Fatal("Expected processed")
	}

	rec, _ := store.GetService(ctx, "svc-1")
	if rec.Version != 1 {
		t.Errorf("Expected version 1, got %d", rec.Version)
	}

	// Second call - Version 2
	processed, err = reg.Reconcile(ctx)
	if err != nil {
		t.Fatalf("Reconcile 2 failed: %v", err)
	}
	if !processed {
		t.Fatal("Expected processed")
	}

	rec, _ = store.GetService(ctx, "svc-1")
	if rec.Version != 2 {
		t.Errorf("Expected version 2, got %d", rec.Version)
	}

	// Verify Runtime: Version 2 should be active, Version 1 unloaded
	if _, ok := runtime.modules["svc-1-2"]; !ok {
		t.Errorf("Service version 2 not found in runtime")
	}
	if _, ok := runtime.modules["svc-1-1"]; ok {
		t.Errorf("Service version 1 should be unloaded")
	}
	if ver, ok := runtime.activeVersions["svc-1"]; !ok || ver != "2" {
		t.Errorf("Service active version mismatch: got %v, want 2", ver)
	}
}
