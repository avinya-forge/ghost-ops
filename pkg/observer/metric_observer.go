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

// Poll is exported for testing purposes.
func (o *MetricObserver) Poll(ctx context.Context) {
	o.poll(ctx)
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

	shadowErrorDelta := make(map[string]int64)
	shadowSuccessDelta := make(map[string]int64)

	for key, value := range metrics {
		if valInt, ok := value.(int64); ok {
			// check for invoke_error and invoke_success
			if strings.HasPrefix(key, "invoke_error{") {
				serviceID := extractServiceID(key)
				if serviceID != "" {
					delta := valInt - o.prevCounters[key]
					if strings.Contains(key, "type=shadow") {
						shadowErrorDelta[serviceID] += delta
					} else {
						servicesErrorDelta[serviceID] += delta
					}
					o.prevCounters[key] = valInt
				}
			} else if strings.HasPrefix(key, "invoke_success{") {
				serviceID := extractServiceID(key)
				if serviceID != "" {
					delta := valInt - o.prevCounters[key]
					if strings.Contains(key, "type=shadow") {
						shadowSuccessDelta[serviceID] += delta
					} else {
						servicesSuccessDelta[serviceID] += delta
					}
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

	// Calculate shadow vs active metrics (Task 505, 506, 507)
	// Iterate through all services that had shadow metrics (success or error)
	shadowServices := make(map[string]bool)
	for id := range shadowErrorDelta {
		shadowServices[id] = true
	}
	for id := range shadowSuccessDelta {
		shadowServices[id] = true
	}

	for serviceID := range shadowServices {
		shadowErrs := shadowErrorDelta[serviceID]
		shadowSuccs := shadowSuccessDelta[serviceID]
		shadowTotal := shadowErrs + shadowSuccs

		// We need enough data to make a confident decision
		if shadowTotal > 5 {
			shadowErrorRate := float64(shadowErrs) / float64(shadowTotal)

			// Get active error rate for the same interval
			activeErrs := servicesErrorDelta[serviceID]
			activeSuccs := servicesSuccessDelta[serviceID]
			activeTotal := activeErrs + activeSuccs

			activeErrorRate := 0.0
			if activeTotal > 0 {
				activeErrorRate = float64(activeErrs) / float64(activeTotal)
			} else {
				// If no active traffic in this interval, use a slightly higher threshold to ensure it's at least better than default threshold
				activeErrorRate = 0.01 // Default acceptable error rate
			}

			if shadowErrorRate <= activeErrorRate {
				// Shadow is better or equal -> PROMOTE
				o.publishPromotion(ctx, serviceID, "Shadow error rate meets or exceeds active", shadowErrorRate, activeErrorRate)
			} else {
				// Shadow is worse -> ROLLBACK
				o.publishRollback(ctx, serviceID, "Shadow error rate is worse than active", shadowErrorRate, activeErrorRate)
			}
		}
	}
}

func (o *MetricObserver) publishPromotion(ctx context.Context, serviceID, reason string, shadowValue, activeValue float64) {
	if o.eventBus == nil {
		return
	}
	event := protocol.Event{
		Type:      protocol.EventPromotionRequired,
		ServiceID: serviceID,
		Payload: map[string]interface{}{
			"reason":       reason,
			"shadow_error": fmt.Sprintf("%.4f", shadowValue),
			"active_error": fmt.Sprintf("%.4f", activeValue),
		},
		Timestamp: time.Now().UnixNano(),
	}
	_ = o.eventBus.Publish(ctx, event)
}

func (o *MetricObserver) publishRollback(ctx context.Context, serviceID, reason string, shadowValue, activeValue float64) {
	if o.eventBus == nil {
		return
	}
	event := protocol.Event{
		Type:      protocol.EventRollbackRequired,
		ServiceID: serviceID,
		Payload: map[string]interface{}{
			"reason":       reason,
			"shadow_error": fmt.Sprintf("%.4f", shadowValue),
			"active_error": fmt.Sprintf("%.4f", activeValue),
		},
		Timestamp: time.Now().UnixNano(),
	}
	_ = o.eventBus.Publish(ctx, event)
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
