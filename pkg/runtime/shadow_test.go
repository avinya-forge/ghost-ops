package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/protocol"
)

// MockMetricsCollector captures metric calls for verification.
type MockMetricsCollector struct {
	mu         sync.Mutex
	counters   map[string]int64
	gauges     map[string]float64
	histograms map[string][]float64
}

func NewMockMetricsCollector() *MockMetricsCollector {
	return &MockMetricsCollector{
		counters:   make(map[string]int64),
		gauges:     make(map[string]float64),
		histograms: make(map[string][]float64),
	}
}

func (m *MockMetricsCollector) Counter(name string, value int64, labels map[string]string) {
	key := m.makeKey(name, labels)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.counters == nil {
		m.counters = make(map[string]int64)
	}
	m.counters[key] += value
}

func (m *MockMetricsCollector) Gauge(name string, value float64, labels map[string]string) {
	key := m.makeKey(name, labels)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.gauges == nil {
		m.gauges = make(map[string]float64)
	}
	m.gauges[key] = value
}

func (m *MockMetricsCollector) Histogram(name string, value float64, labels map[string]string) {
	key := m.makeKey(name, labels)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.histograms == nil {
		m.histograms = make(map[string][]float64)
	}
	m.histograms[key] = append(m.histograms[key], value)
}

func (m *MockMetricsCollector) makeKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(parts)
	return fmt.Sprintf("%s{%s}", name, strings.Join(parts, ","))
}

// GetCounter returns the value of a specific counter by full key constructed from name and labels
func (m *MockMetricsCollector) GetCounter(name string, labels map[string]string) int64 {
	key := m.makeKey(name, labels)
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.counters[key]
}

func TestWazeroRuntimeHost_ShadowMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	store := protocol.NewInMemoryStateStore()
	collector := NewMockMetricsCollector()

	host, err := NewWazeroRuntimeHost(ctx, store, collector)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// Compile kv-service
	engine := evolution.NewGoCompilerEngine()
	// Resolves to repo root/examples/services/kv-service/main.go
	// Assuming running from pkg/runtime
	srcPath, err := filepath.Abs("../../examples/services/kv-service/main.go")
	if err != nil {
		t.Fatalf("Failed to resolve source path: %v", err)
	}
	if _, err := os.Stat(srcPath); err != nil {
		t.Fatalf("Source file not found at %s: %v", srcPath, err)
	}

	bp := protocol.Blueprint{
		Constraints: map[string]interface{}{
			"source_path": srcPath,
		},
	}

	wasmBytes, err := engine.Evolve(ctx, bp)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	serviceID := "shadow-test-service"
	activeVer := "v1"
	shadowVer := "v2"

	// 1. Load Active Version
	if err := host.LoadModule(ctx, serviceID, activeVer, wasmBytes); err != nil {
		t.Fatalf("Failed to load active module: %v", err)
	}
	if err := host.SetActiveVersion(ctx, serviceID, activeVer); err != nil {
		t.Fatalf("Failed to set active version: %v", err)
	}

	// 2. Load Shadow Version
	if err := host.LoadModule(ctx, serviceID, shadowVer, wasmBytes); err != nil {
		t.Fatalf("Failed to load shadow module: %v", err)
	}
	if err := host.SetShadowVersion(ctx, serviceID, shadowVer); err != nil {
		t.Fatalf("Failed to set shadow version: %v", err)
	}

	// Setup data
	if err := store.Set(ctx, "shadow_key", []byte("shadow_value")); err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// 3. Invoke Service
	payload := []byte("shadow_key")
	// No need to sleep; Invoke will block until the module is ready and listening on the channel

	result, err := host.Invoke(ctx, serviceID, "Handle", payload)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	expected := "value: shadow_value"
	if string(result) != expected {
		t.Errorf("Expected active result %q, got %q", expected, string(result))
	}

	// 4. Verify Metrics
	// Verify Active Success
	activeLabels := map[string]string{"service_id": serviceID}
	// "invoke_success" uses type=shadow for shadow, but for active it doesn't specify type?
	// Let's check wazero_host.go:
	// h.collector.Counter("invoke_success", 1, map[string]string{"service_id": serviceID})

	if count := collector.GetCounter("invoke_success", activeLabels); count != 1 {
		t.Errorf("Expected 1 active invocation, got %d", count)
	}

	// Verify Shadow Success (Async)
	shadowLabels := map[string]string{"service_id": serviceID, "type": "shadow"}

	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	success := false
Loop:
	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for shadow invocation metric")
		case <-ticker.C:
			count := collector.GetCounter("invoke_success", shadowLabels)
			if count == 1 {
				success = true
				break Loop
			}
		}
	}

	if !success {
		t.Error("Shadow invocation metric not found")
	}
}
