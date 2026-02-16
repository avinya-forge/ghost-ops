package protocol

// MetricsCollector defines the interface for collecting runtime metrics.
type MetricsCollector interface {
	// Counter increments a counter metric.
	Counter(name string, value int64, labels map[string]string)

	// Gauge sets a gauge metric.
	Gauge(name string, value float64, labels map[string]string)

	// Histogram records a value in a histogram.
	Histogram(name string, value float64, labels map[string]string)
}
