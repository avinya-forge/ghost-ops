package evolution

import (
	"testing"

	"ghost-ops/pkg/protocol"
)

func TestPriorityQueue_PushPopOrder(t *testing.T) {
	pq := NewPriorityQueue()

	// Push with different priorities
	pq.Push(protocol.Blueprint{ServiceID: "svc-low"}, 10)
	pq.Push(protocol.Blueprint{ServiceID: "svc-high"}, 1)
	pq.Push(protocol.Blueprint{ServiceID: "svc-med"}, 5)

	if pq.Len() != 3 {
		t.Fatalf("Expected queue length 3, got %d", pq.Len())
	}

	// Pop order should be high (1) -> med (5) -> low (10)
	t1, ok := pq.Pop()
	if !ok || t1.Blueprint.ServiceID != "svc-high" || t1.Priority != 1 {
		t.Errorf("Expected first pop to be svc-high (priority 1), got %v", t1)
	}

	t2, ok := pq.Pop()
	if !ok || t2.Blueprint.ServiceID != "svc-med" || t2.Priority != 5 {
		t.Errorf("Expected second pop to be svc-med (priority 5), got %v", t2)
	}

	t3, ok := pq.Pop()
	if !ok || t3.Blueprint.ServiceID != "svc-low" || t3.Priority != 10 {
		t.Errorf("Expected third pop to be svc-low (priority 10), got %v", t3)
	}

	// Queue should be empty now
	if pq.Len() != 0 {
		t.Errorf("Expected queue length 0, got %d", pq.Len())
	}

	// Try popping from empty queue
	_, ok = pq.Pop()
	if ok {
		t.Error("Expected pop from empty queue to return false")
	}
}
