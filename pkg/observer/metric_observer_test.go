package observer

import (
	"context"
	"testing"
	"time"

	"ghost-ops/pkg/protocol"
)

type MockSnapshotter struct {
	MockSnapshot map[string]interface{}
}

func (m *MockSnapshotter) Snapshot() map[string]interface{} {
	return m.MockSnapshot
}

func (m *MockSnapshotter) Counter(name string, value int64, tags map[string]string)     {}
func (m *MockSnapshotter) Gauge(name string, value float64, tags map[string]string)     {}
func (m *MockSnapshotter) Histogram(name string, value float64, tags map[string]string) {}

func TestExtractServiceID(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"invoke_duration_seconds{method=default,service_id=svc-1}", "svc-1"},
		{"invoke_error{service_id=my-service,type=execution_error}", "my-service"},
		{"invoke_success{service_id=another-svc}", "another-svc"},
		{"some_metric{no_service_id=true}", ""},
		{"bad_format", ""},
	}

	for _, tt := range tests {
		actual := extractServiceID(tt.key)
		if actual != tt.expected {
			t.Errorf("extractServiceID(%q) = %q; want %q", tt.key, actual, tt.expected)
		}
	}
}

func TestMetricObserver_AnalyzeMetrics_LatencySpike(t *testing.T) {
	ctx := context.Background()
	bus := &MockEventBus{}
	collector := &MockSnapshotter{
		MockSnapshot: map[string]interface{}{
			"invoke_duration_seconds{service_id=svc-latency}": map[string]interface{}{
				"sum":   1.5,
				"count": int64(2),
				"avg":   0.75,
				"p99":   0.6, // greater than 0.5 threshold
			},
		},
	}

	obs := NewMetricObserver(collector, bus)
	obs.poll(ctx)

	if len(bus.Published) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(bus.Published))
	}

	event := bus.Published[0]
	if event.Type != protocol.EventRePromptRequired {
		t.Errorf("Expected event type %s, got %s", protocol.EventRePromptRequired, event.Type)
	}
	if event.ServiceID != "svc-latency" {
		t.Errorf("Expected service ID svc-latency, got %s", event.ServiceID)
	}
	if event.Payload["reason"] != "Latency spike" {
		t.Errorf("Expected reason 'Latency spike', got %v", event.Payload["reason"])
	}
}

func TestMetricObserver_AnalyzeMetrics_HighErrorRate(t *testing.T) {
	ctx := context.Background()
	bus := &MockEventBus{}
	collector := &MockSnapshotter{
		MockSnapshot: map[string]interface{}{
			"invoke_error{service_id=svc-errors,type=execution_error}": int64(2),
			"invoke_success{service_id=svc-errors}":                    int64(8), // total 10 > 5 threshold, rate 20% > 1%
		},
	}

	obs := NewMetricObserver(collector, bus)
	obs.poll(ctx)

	if len(bus.Published) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(bus.Published))
	}

	event := bus.Published[0]
	if event.Type != protocol.EventRePromptRequired {
		t.Errorf("Expected event type %s, got %s", protocol.EventRePromptRequired, event.Type)
	}
	if event.ServiceID != "svc-errors" {
		t.Errorf("Expected service ID svc-errors, got %s", event.ServiceID)
	}
	if event.Payload["reason"] != "High error rate" {
		t.Errorf("Expected reason 'High error rate', got %v", event.Payload["reason"])
	}

	// Poll again with same snapshot should NOT trigger another event, because delta is 0
	bus.Published = nil
	obs.poll(ctx)
	if len(bus.Published) != 0 {
		t.Errorf("Expected 0 events on second poll with no delta, got %d", len(bus.Published))
	}
}

func TestMetricObserver_AnalyzeMetrics_Promotion(t *testing.T) {
	ctx := context.Background()
	bus := &MockEventBus{}
	collector := &MockSnapshotter{
		MockSnapshot: map[string]interface{}{
			// Active errors = 2, success = 18 -> Total 20 -> Error rate 10%
			"invoke_error{service_id=svc-promote,type=execution_error}": int64(2),
			"invoke_success{service_id=svc-promote}":                    int64(18),
			// Shadow errors = 0, success = 6 -> Total 6 -> Error rate 0%
			"invoke_error{service_id=svc-promote,type=shadow}":   int64(0),
			"invoke_success{service_id=svc-promote,type=shadow}": int64(6),
		},
	}

	obs := NewMetricObserver(collector, bus)
	obs.poll(ctx)

	var promoteEvent *protocol.Event
	for _, e := range bus.Published {
		if e.Type == protocol.EventPromotionRequired {
			promoteEvent = &e
			break
		}
	}

	if promoteEvent == nil {
		t.Fatalf("Expected EventPromotionRequired, but none published")
	}

	if promoteEvent.ServiceID != "svc-promote" {
		t.Errorf("Expected service ID svc-promote, got %s", promoteEvent.ServiceID)
	}
	if promoteEvent.Payload["shadow_error"] != "0.0000" {
		t.Errorf("Expected shadow error 0.0000, got %v", promoteEvent.Payload["shadow_error"])
	}
	if promoteEvent.Payload["active_error"] != "0.1000" {
		t.Errorf("Expected active error 0.1000, got %v", promoteEvent.Payload["active_error"])
	}
}

func TestMetricObserver_AnalyzeMetrics_Rollback(t *testing.T) {
	ctx := context.Background()
	bus := &MockEventBus{}
	collector := &MockSnapshotter{
		MockSnapshot: map[string]interface{}{
			// Active errors = 0, success = 20 -> Total 20 -> Error rate 0%
			"invoke_error{service_id=svc-rollback,type=execution_error}": int64(0),
			"invoke_success{service_id=svc-rollback}":                    int64(20),
			// Shadow errors = 2, success = 4 -> Total 6 -> Error rate 33.33%
			"invoke_error{service_id=svc-rollback,type=shadow_execution_error}": int64(2),
			"invoke_success{service_id=svc-rollback,type=shadow}":               int64(4),
		},
	}

	obs := NewMetricObserver(collector, bus)
	obs.poll(ctx)

	var rollbackEvent *protocol.Event
	for _, e := range bus.Published {
		if e.Type == protocol.EventRollbackRequired {
			rollbackEvent = &e
			break
		}
	}

	if rollbackEvent == nil {
		t.Fatalf("Expected EventRollbackRequired, but none published")
	}

	if rollbackEvent.ServiceID != "svc-rollback" {
		t.Errorf("Expected service ID svc-rollback, got %s", rollbackEvent.ServiceID)
	}
	if rollbackEvent.Payload["shadow_error"] != "0.3333" {
		t.Errorf("Expected shadow error 0.3333, got %v", rollbackEvent.Payload["shadow_error"])
	}
	if rollbackEvent.Payload["active_error"] != "0.0000" {
		t.Errorf("Expected active error 0.0000, got %v", rollbackEvent.Payload["active_error"])
	}
}

func TestMetricObserver_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	bus := &MockEventBus{}
	collector := &MockSnapshotter{
		MockSnapshot: map[string]interface{}{},
	}

	obs := NewMetricObserver(collector, bus)
	obs.Start(ctx, 10*time.Millisecond)

	time.Sleep(50 * time.Millisecond)
	cancel() // stop observer

	// wait for goroutine to exit
	time.Sleep(50 * time.Millisecond)

	if len(bus.Published) != 0 {
		t.Errorf("Expected 0 events, got %d", len(bus.Published))
	}
}
