package observer

import (
	"context"
	"testing"
	"time"

	"ghost-ops/pkg/protocol"
)

// MockLLMProvider for testing
type MockLLMProvider struct {
	Response string
	Err      error
}

func (m *MockLLMProvider) GenerateCode(ctx context.Context, blueprint protocol.Blueprint) (string, *protocol.TokenUsage, error) {
	return m.Response, &protocol.TokenUsage{}, m.Err
}

// MockEventBus for testing
type MockEventBus struct {
	Published []protocol.Event
}

func (m *MockEventBus) Publish(ctx context.Context, event protocol.Event) error {
	m.Published = append(m.Published, event)
	return nil
}

func (m *MockEventBus) Subscribe(ctx context.Context, eventType protocol.EventType) (<-chan protocol.Event, error) {
	return nil, nil
}

// MockMetricsCollector for testing
type MockMetricsCollector struct {
	Counters map[string]int64
}

func (m *MockMetricsCollector) Counter(name string, value int64, tags map[string]string) {
	if m.Counters == nil {
		m.Counters = make(map[string]int64)
	}
	m.Counters[name] += value
}

func (m *MockMetricsCollector) Gauge(name string, value float64, tags map[string]string) {}
func (m *MockMetricsCollector) Histogram(name string, value float64, tags map[string]string) {}

func TestLogObserver_Sanitize(t *testing.T) {
	obs := NewLogObserver(nil, nil, nil, 10, time.Second)

	tests := []struct {
		input    string
		expected string
	}{
		{"Normal log line", "Normal log line"},
		{"User email: test@example.com", "User email: [REDACTED_EMAIL]"},
		{"Connecting with password=supersecret123 to DB", "Connecting with password=[REDACTED] to DB"},
		{"API token: abcdef123456", "API token:[REDACTED]"},
		{"Using secret key", "Using [REDACTED_SECRET] key"},
		{"No match but has an @ sign", "No match but has an @ sign"},
	}

	for _, tt := range tests {
		actual := obs.sanitize(tt.input)
		if actual != tt.expected {
			t.Errorf("sanitize(%q) = %q; want %q", tt.input, actual, tt.expected)
		}
	}
}

func TestLogObserver_ObserveAndFlush(t *testing.T) {
	ctx := context.Background()
	llm := &MockLLMProvider{Response: `{"anomaly_detected": false, "reason": ""}`}
	bus := &MockEventBus{}
	collector := &MockMetricsCollector{}

	obs := NewLogObserver(llm, bus, collector, 2, time.Second)

	obs.Observe(ctx, "svc-1", "log line 1")
	if len(bus.Published) != 0 {
		t.Errorf("Expected 0 published events, got %d", len(bus.Published))
	}
	if collector.Counters["logs_observed_total"] != 1 {
		t.Errorf("Expected 1 log observed, got %d", collector.Counters["logs_observed_total"])
	}

	obs.Observe(ctx, "svc-1", "log line 2")
	// Flush should have happened because batch size is 2, asynchronously
	time.Sleep(50 * time.Millisecond)

	// AnalyzeLogs should be called, but LLM mock returns anomaly_detected: false
	if len(bus.Published) != 0 {
		t.Errorf("Expected 0 published events, got %d", len(bus.Published))
	}

	obs.mu.Lock()
	if len(obs.buffers["svc-1"]) != 0 {
		t.Errorf("Expected buffer to be empty, got %v", obs.buffers["svc-1"])
	}
	obs.mu.Unlock()
}

func TestLogObserver_AnalyzeLogs_AnomalyDetected(t *testing.T) {
	ctx := context.Background()
	llm := &MockLLMProvider{Response: `{"anomaly_detected": true, "reason": "High error rate"}`}
	bus := &MockEventBus{}

	obs := NewLogObserver(llm, bus, nil, 10, time.Second)

	batch := []string{"error 1", "error 2"}
	obs.AnalyzeLogs(ctx, "svc-anomaly", batch)

	if len(bus.Published) != 1 {
		t.Fatalf("Expected 1 published event, got %d", len(bus.Published))
	}

	event := bus.Published[0]
	if event.Type != protocol.EventAnomalyDetected {
		t.Errorf("Expected event type %s, got %s", protocol.EventAnomalyDetected, event.Type)
	}
	if event.ServiceID != "svc-anomaly" {
		t.Errorf("Expected service ID %s, got %s", "svc-anomaly", event.ServiceID)
	}
	if event.Payload["reason"] != "High error rate" {
		t.Errorf("Expected reason %s, got %s", "High error rate", event.Payload["reason"])
	}
}

func TestLogObserver_AnalyzeLogs_Error(t *testing.T) {
	ctx := context.Background()
	llm := &MockLLMProvider{Response: ``, Err: context.DeadlineExceeded}
	bus := &MockEventBus{}

	obs := NewLogObserver(llm, bus, nil, 10, time.Second)

	batch := []string{"error 1", "error 2"}
	obs.AnalyzeLogs(ctx, "svc-anomaly", batch)

	if len(bus.Published) != 0 {
		t.Fatalf("Expected 0 published events on LLM error, got %d", len(bus.Published))
	}
}

func TestLogObserver_AnalyzeLogs_BadJSON(t *testing.T) {
	ctx := context.Background()
	llm := &MockLLMProvider{Response: `not a json object`}
	bus := &MockEventBus{}

	obs := NewLogObserver(llm, bus, nil, 10, time.Second)

	batch := []string{"error 1", "error 2"}
	obs.AnalyzeLogs(ctx, "svc-anomaly", batch)

	if len(bus.Published) != 0 {
		t.Fatalf("Expected 0 published events on bad JSON, got %d", len(bus.Published))
	}
}

func TestLogObserver_AnalyzeLogs_NoLLM(t *testing.T) {
	ctx := context.Background()
	bus := &MockEventBus{}

	obs := NewLogObserver(nil, bus, nil, 10, time.Second)

	batch := []string{"error 1", "error 2"}
	obs.AnalyzeLogs(ctx, "svc-anomaly", batch)

	if len(bus.Published) != 0 {
		t.Fatalf("Expected 0 published events with no LLM, got %d", len(bus.Published))
	}
}

func TestLogObserver_AnalyzeLogs_MarkdownFormat(t *testing.T) {
	ctx := context.Background()
	llm := &MockLLMProvider{Response: "```json\n{\"anomaly_detected\": true, \"reason\": \"Markdown parsing test\"}\n```"}
	bus := &MockEventBus{}

	obs := NewLogObserver(llm, bus, nil, 10, time.Second)

	batch := []string{"error 1", "error 2"}
	obs.AnalyzeLogs(ctx, "svc-anomaly", batch)

	if len(bus.Published) != 1 {
		t.Fatalf("Expected 1 published event, got %d", len(bus.Published))
	}

	if bus.Published[0].Payload["reason"] != "Markdown parsing test" {
		t.Errorf("Expected markdown to be stripped correctly, got reason: %v", bus.Published[0].Payload["reason"])
	}
}

func TestLogObserver_Observe_EmptySanitized(t *testing.T) {
	ctx := context.Background()
	bus := &MockEventBus{}
	collector := &MockMetricsCollector{}

	obs := NewLogObserver(nil, bus, collector, 10, time.Second)

	obs.Observe(ctx, "svc-1", "")

	obs.mu.Lock()
	if len(obs.buffers["svc-1"]) != 0 {
		t.Errorf("Expected buffer to be empty")
	}
	obs.mu.Unlock()
}

func TestLogObserver_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	llm := &MockLLMProvider{Response: `{"anomaly_detected": false, "reason": ""}`}

	// Use small interval to trigger flush
	obs := NewLogObserver(llm, nil, nil, 10, 10*time.Millisecond)

	obs.Observe(ctx, "svc-1", "log line")

	obs.Start(ctx)

	time.Sleep(50 * time.Millisecond)
	cancel() // Stop the observer

	// Give it a moment to run shutdown flush
	time.Sleep(50 * time.Millisecond)

	obs.mu.Lock()
	defer obs.mu.Unlock()
	if len(obs.buffers["svc-1"]) != 0 {
		t.Errorf("Expected buffer to be flushed, got %v", obs.buffers["svc-1"])
	}
}
