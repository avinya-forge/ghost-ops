package event

import (
	"context"
	"sync"
	"testing"
	"time"

	"ghost-ops/pkg/protocol"
)

func TestInMemoryEventBus(t *testing.T) {
	bus := NewInMemoryEventBus()
	ctx := context.Background()

	// 1. Subscribe
	ch, err := bus.Subscribe(ctx, protocol.EventServiceDeployed)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// 2. Publish
	event := protocol.Event{
		Type:      protocol.EventServiceDeployed,
		ServiceID: "test-service",
		Payload:   map[string]interface{}{"version": 1},
	}

	go func() {
		err := bus.Publish(ctx, event)
		if err != nil {
			t.Errorf("Publish failed: %v", err)
		}
	}()

	// 3. Receive
	select {
	case evt := <-ch:
		if evt.Type != protocol.EventServiceDeployed {
			t.Errorf("Expected event type %s, got %s", protocol.EventServiceDeployed, evt.Type)
		}
		if evt.ServiceID != "test-service" {
			t.Errorf("Expected service ID test-service, got %s", evt.ServiceID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

func TestInMemoryEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewInMemoryEventBus()
	ctx := context.Background()

	ch1, _ := bus.Subscribe(ctx, protocol.EventServiceFailed)
	ch2, _ := bus.Subscribe(ctx, protocol.EventServiceFailed)

	event := protocol.Event{
		Type:      protocol.EventServiceFailed,
		ServiceID: "fail-service",
	}

	bus.Publish(ctx, event)

	// Verify both received
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		select {
		case <-ch1:
		case <-time.After(1 * time.Second):
			t.Errorf("Timeout ch1")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case <-ch2:
		case <-time.After(1 * time.Second):
			t.Errorf("Timeout ch2")
		}
	}()

	wg.Wait()
}
