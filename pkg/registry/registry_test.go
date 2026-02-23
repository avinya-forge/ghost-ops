package registry

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ghost-ops/pkg/event"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/telemetry"
)

// MockStateStore
type MockStateStore struct {
	records map[string]protocol.ServiceRecord
	getServiceErr error
	updateServiceErr error
}

func (m *MockStateStore) GetService(ctx context.Context, id string) (*protocol.ServiceRecord, error) {
	if m.getServiceErr != nil {
		return nil, m.getServiceErr
	}
	if r, ok := m.records[id]; ok {
		return &r, nil
	}
	return nil, nil
}
func (m *MockStateStore) UpdateService(ctx context.Context, r protocol.ServiceRecord) error {
	if m.updateServiceErr != nil {
		return m.updateServiceErr
	}
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
	index      int
	err        error
}

func (m *MockIntentSource) GetNextBlueprint(ctx context.Context) (*protocol.Blueprint, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.index >= len(m.blueprints) {
		return nil, nil
	}
	bp := m.blueprints[m.index]
	m.index++
	return &bp, nil
}

// MockEvolutionEngine
type MockEvolutionEngine struct{
	err error
}

func (m *MockEvolutionEngine) Evolve(ctx context.Context, bp protocol.Blueprint) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []byte("mock-wasm"), nil
}

// MockRuntimeHost
type MockRuntimeHost struct {
	modules        map[string][]byte
	activeVersions map[string]string
	shadowVersions map[string]string
	loadErr        error
	checkHealthErr error
}

func (m *MockRuntimeHost) LoadModule(ctx context.Context, id, version string, b []byte) error {
	if m.loadErr != nil {
		return m.loadErr
	}
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
func (m *MockRuntimeHost) SetShadowVersion(ctx context.Context, id, version string) error {
	if m.shadowVersions == nil {
		m.shadowVersions = make(map[string]string)
	}
	m.shadowVersions[id] = version
	return nil
}
func (m *MockRuntimeHost) UnsetShadowVersion(ctx context.Context, id string) error {
	if m.shadowVersions != nil {
		delete(m.shadowVersions, id)
	}
	return nil
}
func (m *MockRuntimeHost) UnloadVersion(ctx context.Context, id, version string) error {
	uniqueName := fmt.Sprintf("%s-%s", id, version)
	if m.modules != nil {
		delete(m.modules, uniqueName)
	}
	return nil
}
func (m *MockRuntimeHost) CheckHealth(ctx context.Context, id string) error {
	if m.checkHealthErr != nil {
		return m.checkHealthErr
	}
	return nil
}
func (m *MockRuntimeHost) GetLogs(ctx context.Context, id string) ([]byte, error) {
	return nil, nil
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
		modules:        make(map[string][]byte),
		activeVersions: make(map[string]string),
	}
	collector := telemetry.NewInMemoryCollector()

	reg := NewRegistry(store, engine, source, runtime, collector, nil)
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
		modules:        make(map[string][]byte),
		activeVersions: make(map[string]string),
	}
	collector := telemetry.NewInMemoryCollector()

	reg := NewRegistry(store, engine, source, runtime, collector, nil)
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

func TestRegistry_Reconcile_ShadowMode(t *testing.T) {
	store := &MockStateStore{records: make(map[string]protocol.ServiceRecord)}
	engine := &MockEvolutionEngine{}
	source := &MockIntentSource{
		blueprints: []protocol.Blueprint{
			{ServiceID: "svc-1", Intent: "v1"},
			{ServiceID: "svc-1", Intent: "v2", Constraints: map[string]interface{}{"shadow_mode": true}},
		},
	}
	runtime := &MockRuntimeHost{
		modules:        make(map[string][]byte),
		activeVersions: make(map[string]string),
		shadowVersions: make(map[string]string),
	}
	collector := telemetry.NewInMemoryCollector()

	reg := NewRegistry(store, engine, source, runtime, collector, nil)
	ctx := context.Background()

	// 1. Deploy Active (v1)
	processed, err := reg.Reconcile(ctx)
	if err != nil {
		t.Fatalf("Reconcile 1 failed: %v", err)
	}
	if !processed {
		t.Fatal("Expected processed")
	}

	rec, _ := store.GetService(ctx, "svc-1")
	if rec.ActiveVersion != 1 {
		t.Errorf("Expected ActiveVersion 1, got %d", rec.ActiveVersion)
	}
	if rec.ShadowVersion != 0 {
		t.Errorf("Expected ShadowVersion 0, got %d", rec.ShadowVersion)
	}
	if rec.CurrentState != protocol.StateActive {
		t.Errorf("Expected StateActive, got %s", rec.CurrentState)
	}

	// 2. Deploy Shadow (v2)
	processed, err = reg.Reconcile(ctx)
	if err != nil {
		t.Fatalf("Reconcile 2 failed: %v", err)
	}
	if !processed {
		t.Fatal("Expected processed")
	}

	rec, _ = store.GetService(ctx, "svc-1")
	if rec.ActiveVersion != 1 {
		t.Errorf("Expected ActiveVersion 1 to persist, got %d", rec.ActiveVersion)
	}
	if rec.ShadowVersion != 2 {
		t.Errorf("Expected ShadowVersion 2, got %d", rec.ShadowVersion)
	}
	if rec.CurrentState != protocol.StateShadow {
		t.Errorf("Expected StateShadow, got %s", rec.CurrentState)
	}

	// Verify Runtime
	if ver, ok := runtime.activeVersions["svc-1"]; !ok || ver != "1" {
		t.Errorf("Runtime active version mismatch: got %v, want 1", ver)
	}
	if ver, ok := runtime.shadowVersions["svc-1"]; !ok || ver != "2" {
		t.Errorf("Runtime shadow version mismatch: got %v, want 2", ver)
	}

	// Verify both modules are loaded
	if _, ok := runtime.modules["svc-1-1"]; !ok {
		t.Error("v1 module missing")
	}
	if _, ok := runtime.modules["svc-1-2"]; !ok {
		t.Error("v2 module missing")
	}
}

func TestRegistry_Reconcile_Errors(t *testing.T) {
	ctx := context.Background()
	collector := telemetry.NewInMemoryCollector()

	// 1. Intent Source Error
	t.Run("IntentSource Error", func(t *testing.T) {
		source := &MockIntentSource{err: fmt.Errorf("intent error")}
		reg := NewRegistry(nil, nil, source, nil, collector, nil)
		_, err := reg.Reconcile(ctx)
		if err == nil {
			t.Error("Expected error from intent source, got nil")
		}
	})

	// 2. Store GetService Error
	t.Run("Store GetService Error", func(t *testing.T) {
		source := &MockIntentSource{
			blueprints: []protocol.Blueprint{{ServiceID: "svc-1"}},
		}
		store := &MockStateStore{getServiceErr: fmt.Errorf("store error")}
		engine := &MockEvolutionEngine{}
		reg := NewRegistry(store, engine, source, nil, collector, nil)
		// It should NOT return error from Reconcile, but log error and continue
		// Actually Reconcile returns (bool, error). The error is usually nil unless critical.
		// Failures in evolve/store/runtime usually return (true, nil) but increment metrics/log.
		// Let's check if it returns.
		_, _ = reg.Reconcile(ctx)
		// We can't easily check for internal error logging without a better collector or logger hook.
		// But at least it shouldn't panic.
	})

	// 3. Evolution Engine Error
	t.Run("Evolution Engine Error", func(t *testing.T) {
		source := &MockIntentSource{
			blueprints: []protocol.Blueprint{{ServiceID: "svc-1"}},
		}
		store := &MockStateStore{records: make(map[string]protocol.ServiceRecord)}
		engine := &MockEvolutionEngine{err: fmt.Errorf("compile error")}
		reg := NewRegistry(store, engine, source, nil, collector, nil)
		_, err := reg.Reconcile(ctx)
		// Again, Reconcile swallows the error and logs it.
		if err != nil {
			t.Errorf("Expected nil error (swallowed), got %v", err)
		}
	})

	// 4. Runtime LoadModule Error
	t.Run("Runtime LoadModule Error", func(t *testing.T) {
		source := &MockIntentSource{
			blueprints: []protocol.Blueprint{{ServiceID: "svc-1"}},
		}
		store := &MockStateStore{records: make(map[string]protocol.ServiceRecord)}
		engine := &MockEvolutionEngine{}
		runtime := &MockRuntimeHost{loadErr: fmt.Errorf("runtime error")}
		reg := NewRegistry(store, engine, source, runtime, collector, nil)
		_, err := reg.Reconcile(ctx)
		// Reconcile swallows runtime errors too
		if err != nil {
			t.Errorf("Expected nil error (swallowed), got %v", err)
		}
	})
}

func TestRegistry_Metrics(t *testing.T) {
	store := &MockStateStore{records: make(map[string]protocol.ServiceRecord)}
	engine := &MockEvolutionEngine{}
	source := &MockIntentSource{
		blueprints: []protocol.Blueprint{
			{ServiceID: "svc-metrics", Intent: "do metrics"},
		},
	}
	runtime := &MockRuntimeHost{
		modules:        make(map[string][]byte),
		activeVersions: make(map[string]string),
	}
	collector := telemetry.NewInMemoryCollector()

	reg := NewRegistry(store, engine, source, runtime, collector, nil)
	ctx := context.Background()

	// Run Reconcile
	_, err := reg.Reconcile(ctx)
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	snapshot := collector.Snapshot()

	// Check histograms
	histograms := []string{
		"reconcile_duration_seconds{service_id=svc-metrics}",
		"evolve_duration_seconds{service_id=svc-metrics}",
		"deploy_duration_seconds{service_id=svc-metrics}",
	}

	for _, name := range histograms {
		if _, ok := snapshot[name]; !ok {
			t.Errorf("Metric %s not found in snapshot", name)
		}
	}
}

func TestRegistry_Events(t *testing.T) {
	store := &MockStateStore{records: make(map[string]protocol.ServiceRecord)}
	engine := &MockEvolutionEngine{}
	source := &MockIntentSource{
		blueprints: []protocol.Blueprint{
			{ServiceID: "svc-events", Intent: "do events"},
		},
	}
	runtime := &MockRuntimeHost{
		modules:        make(map[string][]byte),
		activeVersions: make(map[string]string),
	}
	collector := telemetry.NewInMemoryCollector()
	bus := event.NewInMemoryEventBus()

	reg := NewRegistry(store, engine, source, runtime, collector, bus)
	ctx := context.Background()

	// Subscribe to ServiceDeployed
	ch, err := bus.Subscribe(ctx, protocol.EventServiceDeployed)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Run Reconcile
	_, err = reg.Reconcile(ctx)
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	// Wait for event
	select {
	case evt := <-ch:
		if evt.Type != protocol.EventServiceDeployed {
			t.Errorf("Expected event type %s, got %s", protocol.EventServiceDeployed, evt.Type)
		}
		if evt.ServiceID != "svc-events" {
			t.Errorf("Expected service ID svc-events, got %s", evt.ServiceID)
		}
		if ver, ok := evt.Payload["version"]; !ok || ver != 1 {
			t.Errorf("Expected payload version 1, got %v", ver)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event")
	}
}
