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

func TestPriorityQueue_Update(t *testing.T) {
	pq := NewPriorityQueue()

	task1 := pq.Push(protocol.Blueprint{ServiceID: "svc-low"}, 10)
	_ = pq.Push(protocol.Blueprint{ServiceID: "svc-high"}, 1) // task2
	task3 := pq.Push(protocol.Blueprint{ServiceID: "svc-med"}, 5)

	// Update low priority task to be the highest priority
	newBp := protocol.Blueprint{ServiceID: "svc-low-updated"}
	pq.Update(task1, newBp, 0)

	// Update another task just to test index shifts
	pq.Update(task3, protocol.Blueprint{ServiceID: "svc-med-updated"}, 2)

	// The order should now be: task1 (0) -> task2 (1) -> task3 (2)
	t1, ok := pq.Pop()
	if !ok || t1.Blueprint.ServiceID != "svc-low-updated" || t1.Priority != 0 {
		t.Errorf("Expected first pop to be updated task1 (priority 0), got %v", t1)
	}

	t2, ok := pq.Pop()
	if !ok || t2.Blueprint.ServiceID != "svc-high" || t2.Priority != 1 {
		t.Errorf("Expected second pop to be task2 (priority 1), got %v", t2)
	}

	t3, ok := pq.Pop()
	if !ok || t3.Blueprint.ServiceID != "svc-med-updated" || t3.Priority != 2 {
		t.Errorf("Expected third pop to be updated task3 (priority 2), got %v", t3)
	}
}

func TestPriorityQueue_Remove(t *testing.T) {
	pq := NewPriorityQueue()

	task1 := pq.Push(protocol.Blueprint{ServiceID: "svc-low"}, 10)
	task2 := pq.Push(protocol.Blueprint{ServiceID: "svc-high"}, 1)
	pq.Push(protocol.Blueprint{ServiceID: "svc-med"}, 5)

	// Remove highest priority task
	pq.Remove(task2)

	if pq.Len() != 2 {
		t.Fatalf("Expected queue length 2 after removal, got %d", pq.Len())
	}

	// Remove task1 (lowest)
	pq.Remove(task1)

	if pq.Len() != 1 {
		t.Fatalf("Expected queue length 1 after removal, got %d", pq.Len())
	}

	// Try removing task1 again (should be ignored safely)
	pq.Remove(task1)
	if pq.Len() != 1 {
		t.Fatalf("Expected queue length 1 after safe double-removal, got %d", pq.Len())
	}

	// Pop the remaining
	t3, ok := pq.Pop()
	if !ok || t3.Blueprint.ServiceID != "svc-med" || t3.Priority != 5 {
		t.Errorf("Expected pop to be svc-med (priority 5), got %v", t3)
	}
}

func TestPriorityQueue_UpdateInvalid(t *testing.T) {
	pq := NewPriorityQueue()
	task := pq.Push(protocol.Blueprint{ServiceID: "svc1"}, 1)
	pq.Remove(task)
	// Task is removed, Index is -1
	pq.Update(task, protocol.Blueprint{ServiceID: "svc2"}, 2) // Should safely ignore

	if pq.Len() != 0 {
		t.Errorf("Expected length 0, got %d", pq.Len())
	}
}
