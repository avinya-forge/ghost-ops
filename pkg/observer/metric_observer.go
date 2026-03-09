package observer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ghost-ops/pkg/protocol"
)

// MetricObserver continuously monitors runtime metrics and triggers events
// when thresholds (like error rate or latency spikes) are exceeded.
type MetricObserver struct {
	collector protocol.MetricsCollector
	eventBus  protocol.EventBus

	// track previous values to compute delta over the interval
	prevCounters map[string]int64
}

type snapshotter interface {
	Snapshot() map[string]interface{}
}

// NewMetricObserver creates a new MetricObserver.
func NewMetricObserver(collector protocol.MetricsCollector, eventBus protocol.EventBus) *MetricObserver {
	return &MetricObserver{
		collector:    collector,
		eventBus:     eventBus,
		prevCounters: make(map[string]int64),
	}
}

// Start begins the background loop that polls metrics.
func (o *MetricObserver) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				o.poll(ctx)
			}
		}
	}()
}

func (o *MetricObserver) poll(ctx context.Context) {
	snap, ok := o.collector.(snapshotter)
	if !ok {
		return // collector does not support Snapshot()
	}

	metrics := snap.Snapshot()
	o.analyzeMetrics(ctx, metrics)
}

func (o *MetricObserver) analyzeMetrics(ctx context.Context, metrics map[string]interface{}) {
	servicesErrorDelta := make(map[string]int64)
	servicesSuccessDelta := make(map[string]int64)

	for key, value := range metrics {
		if valInt, ok := value.(int64); ok {
			// check for invoke_error and invoke_success
			if strings.HasPrefix(key, "invoke_error{") {
				serviceID := extractServiceID(key)
				if serviceID != "" {
					delta := valInt - o.prevCounters[key]
					servicesErrorDelta[serviceID] += delta
					o.prevCounters[key] = valInt
				}
			} else if strings.HasPrefix(key, "invoke_success{") {
				serviceID := extractServiceID(key)
				if serviceID != "" {
					delta := valInt - o.prevCounters[key]
					servicesSuccessDelta[serviceID] += delta
					o.prevCounters[key] = valInt
				}
			}
		} else if mapVal, ok := value.(map[string]interface{}); ok {
			// Check for histogram with p99
			if strings.HasPrefix(key, "invoke_duration_seconds{") {
				serviceID := extractServiceID(key)
				if serviceID != "" {
					if p99Val, hasP99 := mapVal["p99"]; hasP99 {
						if p99, isFloat := p99Val.(float64); isFloat && p99 > 0.5 { // 500ms
							o.publishReprompt(ctx, serviceID, "Latency spike", "p99_latency", p99, 0.5)
						}
					}
				}
			}
		}
	}

	// Calculate error rates
	for serviceID, errors := range servicesErrorDelta {
		successes := servicesSuccessDelta[serviceID]
		total := errors + successes

		// Only calculate if we have enough traffic to avoid flapping
		if total > 5 {
			errorRate := float64(errors) / float64(total)
			if errorRate > 0.01 { // 1%
				o.publishReprompt(ctx, serviceID, "High error rate", "error_rate", errorRate, 0.01)
			}
		}
	}
}

func (o *MetricObserver) publishReprompt(ctx context.Context, serviceID, reason, metric string, value, threshold float64) {
	if o.eventBus == nil {
		return
	}
	event := protocol.Event{
		Type:      protocol.EventRePromptRequired,
		ServiceID: serviceID,
		Payload: map[string]interface{}{
			"reason":    reason,
			"metric":    metric,
			"value":     fmt.Sprintf("%.4f", value),
			"threshold": fmt.Sprintf("%.4f", threshold),
		},
		Timestamp: time.Now().UnixNano(),
	}
	_ = o.eventBus.Publish(ctx, event)
}

func extractServiceID(key string) string {
	// key format: metric_name{label1=val1,service_id=svc-1,label2=val2}
	start := strings.Index(key, "{")
	end := strings.Index(key, "}")
	if start == -1 || end == -1 || start > end {
		return ""
	}
	labelsStr := key[start+1 : end]
	parts := strings.Split(labelsStr, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 && kv[0] == "service_id" {
			return kv[1]
		}
	}
	return ""
}
