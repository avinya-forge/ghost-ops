package registry

import (
	"context"
	"testing"

	"ghost-ops/pkg/protocol"
)

// MockRuntimeHost is a mock implementation of RuntimeHost for testing.
type MockRuntimeHost struct {
	LoadedModules map[string][]byte
}

func (m *MockRuntimeHost) LoadModule(ctx context.Context, serviceID string, wasmBytes []byte) error {
	if m.LoadedModules == nil {
		m.LoadedModules = make(map[string][]byte)
	}
	m.LoadedModules[serviceID] = wasmBytes
	return nil
}

func (m *MockRuntimeHost) UnloadModule(ctx context.Context, serviceID string) error {
	if m.LoadedModules != nil {
		delete(m.LoadedModules, serviceID)
	}
	return nil
}

func (m *MockRuntimeHost) Invoke(ctx context.Context, serviceID, method string, payload []byte) ([]byte, error) {
	return []byte("mock-result"), nil
}

// MockEvolutionEngine is a mock implementation of EvolutionEngine for testing.
type MockEvolutionEngine struct{}

func (m *MockEvolutionEngine) Evolve(ctx context.Context, blueprint protocol.Blueprint) ([]byte, error) {
	return []byte("mock-wasm-bytes"), nil
}

// MockIntentSource is a local mock for this test
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

// MockStateStore is a local mock for this test
type MockStateStore struct {
	services map[string]protocol.ServiceRecord
}

func NewMockStateStore() *MockStateStore {
	return &MockStateStore{
		services: make(map[string]protocol.ServiceRecord),
	}
}

func (s *MockStateStore) GetService(ctx context.Context, serviceID string) (*protocol.ServiceRecord, error) {
	if r, ok := s.services[serviceID]; ok {
		return &r, nil
	}
	return nil, nil
}

func (s *MockStateStore) UpdateService(ctx context.Context, record protocol.ServiceRecord) error {
	s.services[record.ServiceID] = record
	return nil
}

func (s *MockStateStore) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	var list []protocol.ServiceRecord
	for _, r := range s.services {
		list = append(list, r)
	}
	return list, nil
}

func TestServiceRegistry_Reconcile(t *testing.T) {
	ctx := context.Background()

	// Setup Mocks
	intentSource := &MockIntentSource{
		Blueprints: []protocol.Blueprint{
			{ServiceID: "service-1", Intent: "Test Service Intent"},
		},
	}

	evolutionEngine := &MockEvolutionEngine{}
	runtimeHost := &MockRuntimeHost{}
	stateStore := NewMockStateStore()

	registry := NewServiceRegistry(intentSource, evolutionEngine, runtimeHost, stateStore)

	// Run Reconcile
	processed, err := registry.Reconcile(ctx)
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if !processed {
		t.Error("Expected to process one blueprint, but processed none")
	}

	// Verify Runtime Host
	if len(runtimeHost.LoadedModules) != 1 {
		t.Errorf("Expected 1 module loaded, got %d", len(runtimeHost.LoadedModules))
	}
	if _, ok := runtimeHost.LoadedModules["service-1"]; !ok {
		t.Error("Expected service-1 to be loaded")
	}

	// Verify State Store
	svc, err := stateStore.GetService(ctx, "service-1")
	if err != nil {
		t.Fatalf("Failed to get service from store: %v", err)
	}
	if svc == nil {
		t.Fatal("Expected service to be found in store")
	}
	if svc.ServiceID != "service-1" {
		t.Errorf("Expected ServiceID 'service-1', got '%s'", svc.ServiceID)
	}
	if svc.CurrentState != protocol.StateActive {
		t.Errorf("Expected StateActive, got %s", svc.CurrentState)
	}

	// Run Reconcile again (should return nil as no more blueprints)
	processed, err = registry.Reconcile(ctx)
	if err != nil {
		t.Fatalf("Second Reconcile failed: %v", err)
	}
	if processed {
		t.Error("Expected no blueprint processed, but one was")
	}
}
