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
	closed      bool

	// A channel that gets closed when the bus is closed.
	// Used to signal waiting subscribers to exit if the bus itself closes.
	closeCh chan struct{}
}

// NewInMemoryEventBus creates a new InMemoryEventBus.
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		subscribers: make(map[protocol.EventType][]chan protocol.Event),
		closeCh:     make(chan struct{}),
	}
}

// Publish publishes an event to all subscribers.
func (b *InMemoryEventBus) Publish(ctx context.Context, event protocol.Event) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return nil
	}

	event.Timestamp = time.Now().Unix()

	var chansToNotify []chan protocol.Event
	if chans, ok := b.subscribers[event.Type]; ok {
		// Copy channel slice while holding the lock
		chansToNotify = append(chansToNotify, chans...)
	}
	closeCh := b.closeCh
	b.mu.RUnlock()

	for _, ch := range chansToNotify {
		select {
		case ch <- event:
		case <-ctx.Done():
			return ctx.Err()
		case <-closeCh:
			// Bus was closed, stop publishing
			return nil
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

	// Keep a reference to the close channel
	closeCh := b.closeCh

	// Launch a goroutine to wait for the context to be canceled or the bus to be closed
	go func(targetCh chan protocol.Event) {
		select {
		case <-ctx.Done():
		case <-closeCh:
		}

		b.mu.Lock()
		defer b.mu.Unlock()

		if b.closed {
			return
		}

		if chans, ok := b.subscribers[eventType]; ok {
			var newChans []chan protocol.Event
			for _, existingCh := range chans {
				if existingCh != targetCh {
					newChans = append(newChans, existingCh)
				}
			}
			b.subscribers[eventType] = newChans
		}

		// Note: We don't close the channel to avoid 'send on closed channel' panics
		// when Publish races with cancellation. Receivers know they should stop reading
		// either because their context was canceled or the bus closed.
	}(ch)

	return ch, nil
}

// Close closes all subscriber channels. Subsequent Publish calls are no-ops.
// Calling Close more than once is safe.
func (b *InMemoryEventBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}
	b.closed = true
	close(b.closeCh)

	b.subscribers = make(map[protocol.EventType][]chan protocol.Event)
}
