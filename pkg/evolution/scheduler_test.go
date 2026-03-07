package evolution

import (
	"testing"

	"ghost-ops/pkg/protocol"
)

func TestParseResourceRequestDefaults(t *testing.T) {
	req, err := ParseResourceRequest(protocol.Blueprint{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.CPUMilli != defaultCPUMilli || req.MemoryMB != defaultMemoryMB {
		t.Fatalf("expected defaults (%d,%d), got (%d,%d)", defaultCPUMilli, defaultMemoryMB, req.CPUMilli, req.MemoryMB)
	}
}

func TestParseResourceRequestConstraints(t *testing.T) {
	bp := protocol.Blueprint{Constraints: map[string]interface{}{"cpu_milli": 250.0, "memory_mb": 512}}
	req, err := ParseResourceRequest(bp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.CPUMilli != 250 || req.MemoryMB != 512 {
		t.Fatalf("expected (250,512), got (%d,%d)", req.CPUMilli, req.MemoryMB)
	}
}

func TestParseResourceRequestRejectsInvalid(t *testing.T) {
	_, err := ParseResourceRequest(protocol.Blueprint{Constraints: map[string]interface{}{"cpu_milli": "high"}})
	if err == nil {
		t.Fatal("expected parse error for cpu_milli")
	}

	_, err = ParseResourceRequest(protocol.Blueprint{Constraints: map[string]interface{}{"memory_mb": 0}})
	if err == nil {
		t.Fatal("expected validation error for memory_mb")
	}
}

func TestPackingScore(t *testing.T) {
	fitSmall := PackingScore(ResourceRequest{CPUMilli: 100, MemoryMB: 128}, ResourceCapacity{CPUMilli: 1000, MemoryMB: 1024})
	fitLarge := PackingScore(ResourceRequest{CPUMilli: 700, MemoryMB: 800}, ResourceCapacity{CPUMilli: 1000, MemoryMB: 1024})
	over := PackingScore(ResourceRequest{CPUMilli: 2000, MemoryMB: 1024}, ResourceCapacity{CPUMilli: 1000, MemoryMB: 1024})

	if !(fitSmall < fitLarge && fitLarge < over) {
		t.Fatalf("expected fitSmall < fitLarge < over, got %d, %d, %d", fitSmall, fitLarge, over)
	}
}

func TestResourceSchedulerEnqueueOrder(t *testing.T) {
	s := NewResourceScheduler(nil)
	cap := ResourceCapacity{CPUMilli: 1000, MemoryMB: 1000}

	_, err := s.Enqueue(protocol.Blueprint{ServiceID: "svc-huge", Constraints: map[string]interface{}{"cpu_milli": 1200}}, cap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = s.Enqueue(protocol.Blueprint{ServiceID: "svc-small", Constraints: map[string]interface{}{"cpu_milli": 100}}, cap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	task, ok := s.Queue().Pop()
	if !ok {
		t.Fatal("expected first task")
	}
	if task.Blueprint.ServiceID != "svc-small" {
		t.Fatalf("expected small service first, got %s", task.Blueprint.ServiceID)
	}
}
