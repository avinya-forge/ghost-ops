package telemetry

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"ghost-ops/pkg/protocol"
)

// Ensure InMemoryCollector implements protocol.MetricsCollector.
var _ protocol.MetricsCollector = (*InMemoryCollector)(nil)

// InMemoryCollector implements protocol.MetricsCollector using in-memory storage.
type InMemoryCollector struct {
	counters   sync.Map
	gauges     sync.Map
	histograms sync.Map
}

// NewInMemoryCollector creates a new InMemoryCollector.
func NewInMemoryCollector() *InMemoryCollector {
	return &InMemoryCollector{}
}

func makeKey(name string, labels map[string]string) string {
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

// Counter increments a counter metric.
func (c *InMemoryCollector) Counter(name string, value int64, labels map[string]string) {
	key := makeKey(name, labels)
	// Initialize with a pointer to int64(0) if not present
	val, _ := c.counters.LoadOrStore(key, new(int64))
	atomic.AddInt64(val.(*int64), value)
}

// Gauge sets a gauge metric.
func (c *InMemoryCollector) Gauge(name string, value float64, labels map[string]string) {
	key := makeKey(name, labels)
	c.gauges.Store(key, value)
}

type safeHistogram struct {
	mu    sync.Mutex
	sum   float64
	count int64
}

func (h *safeHistogram) record(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sum += v
	h.count++
}

// Histogram records a value in a histogram.
func (c *InMemoryCollector) Histogram(name string, value float64, labels map[string]string) {
	key := makeKey(name, labels)
	val, _ := c.histograms.LoadOrStore(key, &safeHistogram{})
	hist := val.(*safeHistogram)
	hist.record(value)
}

// Snapshot returns a snapshot of all metrics.
func (c *InMemoryCollector) Snapshot() map[string]interface{} {
	result := make(map[string]interface{})

	c.counters.Range(func(key, value interface{}) bool {
		result[key.(string)] = atomic.LoadInt64(value.(*int64))
		return true
	})

	c.gauges.Range(func(key, value interface{}) bool {
		result[key.(string)] = value
		return true
	})

	c.histograms.Range(func(key, value interface{}) bool {
		hist := value.(*safeHistogram)
		hist.mu.Lock()
		res := map[string]interface{}{
			"sum":   hist.sum,
			"count": hist.count,
			"avg":   0.0,
		}
		if hist.count > 0 {
			res["avg"] = hist.sum / float64(hist.count)
		}
		hist.mu.Unlock()
		result[key.(string)] = res
		return true
	})

	return result
}
