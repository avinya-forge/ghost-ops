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
	mu     sync.Mutex
	sum    float64
	count  int64
	values []float64
}

func (h *safeHistogram) record(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sum += v
	h.count++
	h.values = append(h.values, v)
	if len(h.values) > 1000 {
		// Keep last 1000 items
		h.values = h.values[1:]
	}
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

		p99 := 0.0
		if len(hist.values) > 0 {
			sortedValues := make([]float64, len(hist.values))
			copy(sortedValues, hist.values)
			sort.Float64s(sortedValues)

			p99Index := int(float64(len(sortedValues)) * 0.99)
			// Clamp to valid range: p99Index is always < len here, but
			// the explicit check prevents an out-of-bounds panic if this
			// block is ever reached with an empty slice in the future.
			if p99Index >= len(sortedValues) {
				p99Index = len(sortedValues) - 1
			}
			if p99Index < 0 {
				p99Index = 0
			}
			p99 = sortedValues[p99Index]
		}

		res := map[string]interface{}{
			"sum":   hist.sum,
			"count": hist.count,
			"avg":   0.0,
			"p99":   p99,
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
