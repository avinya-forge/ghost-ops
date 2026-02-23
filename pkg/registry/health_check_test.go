package registry

import (
	"context"
	"fmt"
	"testing"

	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/telemetry"
)

func TestRegistry_CheckServices_Unhealthy(t *testing.T) {
	// Setup Store with one active service
	store := &MockStateStore{
		records: map[string]protocol.ServiceRecord{
			"svc-unhealthy": {
				ServiceID:     "svc-unhealthy",
				ActiveVersion: 1,
				CurrentState:  protocol.StateActive,
			},
		},
	}

	engine := &MockEvolutionEngine{}
	source := &MockIntentSource{}

	// Setup Runtime with the module loaded, but CheckHealth returns error
	runtime := &MockRuntimeHost{
		modules: map[string][]byte{
			"svc-unhealthy-1": []byte("mock-wasm"),
		},
		activeVersions: map[string]string{
			"svc-unhealthy": "svc-unhealthy-1",
		},
		checkHealthErr: fmt.Errorf("connection refused"),
	}

	collector := telemetry.NewInMemoryCollector()

	reg := NewRegistry(store, engine, source, runtime, collector)
	ctx := context.Background()

	// Perform Health Check
	reg.CheckServices(ctx)

	// Verify Metrics: health_check_failure should be 1
	snapshot := collector.Snapshot()
	if val, ok := snapshot["health_check_failure{service_id=svc-unhealthy}"]; !ok || val != int64(1) {
		t.Errorf("Expected health_check_failure metric 1, got %v", val)
	}

	// Verify UnloadVersion was called (module removed from mock runtime)
	if _, ok := runtime.modules["svc-unhealthy-1"]; ok {
		t.Error("Expected module svc-unhealthy-1 to be unloaded after health check failure")
	}

	// Verify Active Version is NOT removed from store.
	// The implementation of CheckServices calls `runtime.UnloadVersion` but does not update the store.
	// The service remains "Active" in the store, but is unloaded from the runtime.
	// Subsequent invocations will fail until the service is redeployed or healed.
}

func TestRegistry_CheckServices_Healthy(t *testing.T) {
	store := &MockStateStore{
		records: map[string]protocol.ServiceRecord{
			"svc-healthy": {
				ServiceID:     "svc-healthy",
				ActiveVersion: 1,
				CurrentState:  protocol.StateActive,
			},
		},
	}
	engine := &MockEvolutionEngine{}
	source := &MockIntentSource{}
	runtime := &MockRuntimeHost{
		modules: map[string][]byte{
			"svc-healthy-1": []byte("mock-wasm"),
		},
		activeVersions: map[string]string{
			"svc-healthy": "svc-healthy-1",
		},
		checkHealthErr: nil, // Healthy
	}
	collector := telemetry.NewInMemoryCollector()

	reg := NewRegistry(store, engine, source, runtime, collector)
	ctx := context.Background()

	reg.CheckServices(ctx)

	snapshot := collector.Snapshot()
	if val, ok := snapshot["health_check_failure{service_id=svc-healthy}"]; ok && val != 0 {
		t.Errorf("Expected 0 health check failures, got %v", val)
	}

	if _, ok := runtime.modules["svc-healthy-1"]; !ok {
		t.Error("Expected module to remain loaded")
	}
}
