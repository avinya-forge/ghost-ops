package protocol

import "context"

// EventType defines the type of event.
type EventType string

const (
	EventServiceDeployed  EventType = "service_deployed"
	EventServiceFailed    EventType = "service_failed"
	EventServiceUnhealthy EventType = "service_unhealthy"
	EventAnomalyDetected   EventType = "anomaly_detected"
	EventRePromptRequired  EventType = "reprompt_required"
	EventPromotionRequired EventType = "promotion_required"
	EventRollbackRequired  EventType = "rollback_required"
)

// Event represents a system event.
type Event struct {
	Type      EventType
	ServiceID string
	Payload   map[string]interface{}
	Timestamp int64
}

// EventBus defines the interface for publishing and subscribing to events.
type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(ctx context.Context, eventType EventType) (<-chan Event, error)
}
