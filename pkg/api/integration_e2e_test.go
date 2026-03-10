package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ghost-ops/pkg/config"
	"ghost-ops/pkg/event"
	"ghost-ops/pkg/observer"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/registry"
	"ghost-ops/pkg/runtime"
	"ghost-ops/pkg/telemetry"
)

// TestE2E_OptimizationLoop simulates Task 509: End-To-End Optimization Test.
// It verifies the full lifecycle:
// 1. Deploy initial version (v1).
// 2. Simulate traffic that causes errors -> Triggers Reprompt event.
// 3. System deploys new version (v2) in shadow mode.
// 4. Simulate shadow traffic that is better than active -> Triggers Promote event.
// 5. System promotes shadow to active.
// 6. Verify v2 is now active.
func TestE2E_OptimizationLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := protocol.NewInMemoryStateStore()
	collector := telemetry.NewInMemoryCollector()
	bus := event.NewInMemoryEventBus()

	// Use Real Wazero Runtime
	rt, err := runtime.NewWazeroRuntimeHost(ctx, store, collector)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer rt.Close(ctx)

	// Setup Mocks
	source := &MockIntentSource{
		blueprints: []protocol.Blueprint{
			{ServiceID: "e2e-opt-service", Intent: "v1"},
		},
	}

	engine := &MockEvolutionEngine{
		wasmBytes: integrationWasmBytes,
	}

	reg := registry.NewRegistry(store, engine, source, rt, collector, bus, config.RegistryConfig{})

	// Start Event Loop for promotions/rollbacks
	reg.StartEventLoop(ctx)

	// Initialize Metric Observer
	metricObs := observer.NewMetricObserver(collector, bus)

	// We'll manually call metricObs.poll() instead of Start() to control timing in the test

	// Subscribe to RePrompt events
	repromptCh, _ := bus.Subscribe(ctx, protocol.EventRePromptRequired)

	// Create API Server
	srv := NewServer(reg, collector, 1024*1024)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// --- STEP 1: Deploy initial version (v1) ---
	resp, err := http.Post(ts.URL+"/reconcile", "application/json", nil)
	if err != nil {
		t.Fatalf("Reconcile request failed: %v", err)
	}
	resp.Body.Close()

	time.Sleep(100 * time.Millisecond)

	rec, _ := store.GetService(ctx, "e2e-opt-service")
	if rec == nil || rec.ActiveVersion != 1 {
		t.Fatalf("Expected v1 to be active, got: %v", rec)
	}

	// --- STEP 2: Simulate errors on v1 to trigger Reprompt ---
	for i := 0; i < 10; i++ {
		collector.Counter("invoke_error", 1, map[string]string{"service_id": "e2e-opt-service", "type": "execution_error"})
	}
	// Add some success so error rate is not 100%, but > 1%
	for i := 0; i < 10; i++ {
		collector.Counter("invoke_success", 1, map[string]string{"service_id": "e2e-opt-service"})
	}

	metricObs.Poll(ctx)

	// Wait for Reprompt Event
	select {
	case evt := <-repromptCh:
		if evt.ServiceID != "e2e-opt-service" {
			t.Fatalf("Expected reprompt for e2e-opt-service, got %s", evt.ServiceID)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("Timeout waiting for reprompt event")
	}

	// --- STEP 3: System deploys v2 in shadow mode ---
	// Add new blueprint with shadow mode to simulate the Re-Prompt handler generating it
	source.blueprints = append(source.blueprints, protocol.Blueprint{
		ServiceID:   "e2e-opt-service",
		Intent:      "v2 optimized",
		Constraints: map[string]interface{}{"shadow_mode": true},
	})

	resp2, _ := http.Post(ts.URL+"/reconcile", "application/json", nil)
	resp2.Body.Close()

	time.Sleep(100 * time.Millisecond)

	rec, _ = store.GetService(ctx, "e2e-opt-service")
	if rec.ActiveVersion != 1 || rec.ShadowVersion != 2 {
		t.Fatalf("Expected active v1 and shadow v2, got active %d, shadow %d", rec.ActiveVersion, rec.ShadowVersion)
	}

	// --- STEP 4 & 5: Simulate shadow traffic better than active -> Promote ---
	// Active still has 50% error rate (1 error, 1 success)
	collector.Counter("invoke_error", 1, map[string]string{"service_id": "e2e-opt-service", "type": "execution_error"})
	collector.Counter("invoke_success", 1, map[string]string{"service_id": "e2e-opt-service"})

	// Shadow has 0% error rate (0 error, 6 success to pass the >5 threshold)
	for i := 0; i < 6; i++ {
		collector.Counter("invoke_success", 1, map[string]string{"service_id": "e2e-opt-service", "type": "shadow"})
	}

	metricObs.Poll(ctx)

	// Wait for EventLoop to process the promotion
	time.Sleep(500 * time.Millisecond)

	// --- STEP 6: Verify v2 is now active ---
	rec, _ = store.GetService(ctx, "e2e-opt-service")
	if rec.ActiveVersion != 2 {
		t.Fatalf("Expected v2 to be promoted to active, got %d", rec.ActiveVersion)
	}
	if rec.ShadowVersion != 0 {
		t.Fatalf("Expected shadow version to be unset, got %d", rec.ShadowVersion)
	}

	// Verify invocation works (this ensures the WASM module is truly active in RuntimeHost)
	payload := []byte("hello optimized")
	req, _ := http.NewRequest("POST", ts.URL+"/services/e2e-opt-service/invoke?method=echo", bytes.NewReader(payload))
	respInvoke, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to invoke: %v", err)
	}
	defer respInvoke.Body.Close()

	body, _ := io.ReadAll(respInvoke.Body)
	if strings.TrimSpace(string(body)) != "hello optimized" {
		t.Fatalf("Unexpected invoke response: %s", body)
	}
}
