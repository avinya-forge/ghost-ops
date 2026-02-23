package event

import (
	"context"
	"sync"
	"time"

	"ghost-ops/pkg/protocol"
)

// InMemoryEventBus implements protocol.EventBus in memory.
type InMemoryEventBus struct {
	mu          sync.RWMutex
	subscribers map[protocol.EventType][]chan protocol.Event
}

// NewInMemoryEventBus creates a new InMemoryEventBus.
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		subscribers: make(map[protocol.EventType][]chan protocol.Event),
	}
}

// Publish publishes an event to all subscribers.
func (b *InMemoryEventBus) Publish(ctx context.Context, event protocol.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	event.Timestamp = time.Now().Unix()

	if chans, ok := b.subscribers[event.Type]; ok {
		for _, ch := range chans {
			// Non-blocking send to avoid blocking the publisher
			select {
			case ch <- event:
			default:
				// Drop event if channel is full
				// Ideally log this
			}
		}
	}
	return nil
}

// Subscribe subscribes to an event type.
func (b *InMemoryEventBus) Subscribe(ctx context.Context, eventType protocol.EventType) (<-chan protocol.Event, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan protocol.Event, 100) // Buffer size 100
	if b.subscribers == nil {
		b.subscribers = make(map[protocol.EventType][]chan protocol.Event)
	}
	b.subscribers[eventType] = append(b.subscribers[eventType], ch)

	return ch, nil
}

// Close closes all channels (optional cleanup)
func (b *InMemoryEventBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, chans := range b.subscribers {
		for _, ch := range chans {
			close(ch)
		}
	}
	b.subscribers = make(map[protocol.EventType][]chan protocol.Event)
}
