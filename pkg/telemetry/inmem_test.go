package telemetry

import (
	"sync"
	"testing"
)

func TestInMemoryCollector_Counter(t *testing.T) {
	c := NewInMemoryCollector()
	labels := map[string]string{"service": "test", "version": "v1"}

	c.Counter("requests", 1, labels)
	c.Counter("requests", 2, labels)

	snapshot := c.Snapshot()
	key := "requests{service=test,version=v1}"

	val, ok := snapshot[key].(int64)
	if !ok {
		t.Fatalf("expected int64 counter value, got %T", snapshot[key])
	}
	if val != 3 {
		t.Errorf("expected counter value 3, got %d", val)
	}
}

func TestInMemoryCollector_Gauge(t *testing.T) {
	c := NewInMemoryCollector()
	labels := map[string]string{"node": "1"}

	c.Gauge("memory_usage", 1024.0, labels)
	c.Gauge("memory_usage", 2048.0, labels)

	snapshot := c.Snapshot()
	key := "memory_usage{node=1}"

	val, ok := snapshot[key].(float64)
	if !ok {
		t.Fatalf("expected float64 gauge value, got %T", snapshot[key])
	}
	if val != 2048.0 {
		t.Errorf("expected gauge value 2048.0, got %f", val)
	}
}

func TestInMemoryCollector_Histogram(t *testing.T) {
	c := NewInMemoryCollector()
	labels := map[string]string{"method": "GET"}

	c.Histogram("latency", 0.1, labels)
	c.Histogram("latency", 0.3, labels)

	snapshot := c.Snapshot()
	key := "latency{method=GET}"

	data, ok := snapshot[key].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map histogram data, got %T", snapshot[key])
	}

	sum, ok := data["sum"].(float64)
	if !ok || sum != 0.4 {
		t.Errorf("expected sum 0.4, got %v", sum)
	}

	count, ok := data["count"].(int64)
	if !ok || count != 2 {
		t.Errorf("expected count 2, got %v", count)
	}

	avg, ok := data["avg"].(float64)
	if !ok || avg != 0.2 {
		t.Errorf("expected avg 0.2, got %v", avg)
	}

	p99, ok := data["p99"].(float64)
	if !ok || p99 != 0.3 {
		t.Errorf("expected p99 0.3, got %v", p99)
	}
}

func TestInMemoryCollector_Concurrency(t *testing.T) {
	c := NewInMemoryCollector()
	var wg sync.WaitGroup
	labels := map[string]string{"id": "concurrent"}

	workers := 10
	iterations := 100

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				c.Counter("concurrent_ops", 1, labels)
			}
		}()
	}
	wg.Wait()

	snapshot := c.Snapshot()
	key := "concurrent_ops{id=concurrent}"

	val, ok := snapshot[key].(int64)
	if !ok {
		t.Fatalf("expected int64 counter value, got %T", snapshot[key])
	}
	expected := int64(workers * iterations)
	if val != expected {
		t.Errorf("expected counter value %d, got %d", expected, val)
	}
}
