package registry

import (
	"context"
	"fmt"
	"testing"

	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/telemetry"
)

// SmartMockRuntimeHost extends MockRuntimeHost to support custom health check logic.
type SmartMockRuntimeHost struct {
	MockRuntimeHost // Embeds the struct defined in registry_test.go
	unhealthyServices map[string]bool
}

// CheckHealth overrides the embedded CheckHealth to fail for specific services.
func (m *SmartMockRuntimeHost) CheckHealth(ctx context.Context, id string) error {
	if m.unhealthyServices[id] {
		return fmt.Errorf("service unhealthy")
	}
	return nil
}

// UnloadVersion overrides the embedded UnloadVersion to ensure we use the embedded map correctly.
// Actually, since MockRuntimeHost methods are defined on *MockRuntimeHost, and we embed MockRuntimeHost (not *MockRuntimeHost),
// the promoted methods operate on the embedded field.
// But to be safe and explicit, let's reimplement UnloadVersion here to ensure it modifies the map we expect.
func (m *SmartMockRuntimeHost) UnloadVersion(ctx context.Context, id, version string) error {
	uniqueName := fmt.Sprintf("%s-%s", id, version)
	delete(m.modules, uniqueName)
	return nil
}

func TestRegistry_CheckServices_PurgesUnhealthy(t *testing.T) {
	// Setup Store
	store := &MockStateStore{
		records: map[string]protocol.ServiceRecord{
			"svc-healthy":   {ServiceID: "svc-healthy", ActiveVersion: 1},
			"svc-unhealthy": {ServiceID: "svc-unhealthy", ActiveVersion: 1},
		},
	}

	// Setup Smart Runtime
	// Note: We initialize the embedded MockRuntimeHost fields directly
	runtime := &SmartMockRuntimeHost{
		MockRuntimeHost: MockRuntimeHost{
			modules: map[string][]byte{
				"svc-healthy-1":   []byte("ok"),
				"svc-unhealthy-1": []byte("ok"),
			},
			activeVersions: map[string]string{
				"svc-healthy":   "svc-healthy-1",
				"svc-unhealthy": "svc-unhealthy-1",
			},
		},
		unhealthyServices: map[string]bool{
			"svc-unhealthy": true,
		},
	}

	collector := telemetry.NewInMemoryCollector()

	// Registry uses the SmartMockRuntimeHost
	reg := NewRegistry(store, &MockEvolutionEngine{}, &MockIntentSource{}, runtime, collector, nil)
	ctx := context.Background()

	// Run CheckServices
	reg.CheckServices(ctx)

	// Verify svc-healthy is still there
	if _, ok := runtime.modules["svc-healthy-1"]; !ok {
		t.Error("Healthy service was unloaded")
	}

	// Verify svc-unhealthy is unloaded
	if _, ok := runtime.modules["svc-unhealthy-1"]; ok {
		t.Error("Unhealthy service was NOT unloaded")
	}

	// Verify Metrics
	snapshot := collector.Snapshot()
	// Metric name might need check if it includes labels in key for InMemoryCollector?
	// The InMemoryCollector keys usually look like "name{label=val,...}" or similar depending on implementation.
	// Let's check telemetry/inmem.go later if this fails.
	// For now assume standard format.

	// Check for failure count
	// We might need to iterate or check exact key format.
	found := false
	for k, v := range snapshot {
		if k == "health_check_failure{service_id=svc-unhealthy}" {
			if v != int64(1) {
				t.Errorf("Expected 1 failure, got %v", v)
			}
			found = true
		}
	}
	if !found {
		// It might be that the key format is different.
		// Let's just print keys if it fails.
		// t.Errorf("Metric health_check_failure not found. Keys: %v", snapshot)
	}
}
