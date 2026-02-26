package intent

import (
	"container/heap"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"ghost-ops/pkg/protocol"
)

func TestPriorityQueue(t *testing.T) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	bp1 := &protocol.Blueprint{ServiceID: "svc-1", Priority: 1}
	bp2 := &protocol.Blueprint{ServiceID: "svc-2", Priority: 10}
	bp3 := &protocol.Blueprint{ServiceID: "svc-3", Priority: 5}

	heap.Push(pq, bp1)
	heap.Push(pq, bp2)
	heap.Push(pq, bp3)

	assert.Equal(t, 3, pq.Len())

	item := heap.Pop(pq).(*protocol.Blueprint)
	assert.Equal(t, "svc-2", item.ServiceID) // Highest priority (10)

	item = heap.Pop(pq).(*protocol.Blueprint)
	assert.Equal(t, "svc-3", item.ServiceID) // Next highest (5)

	item = heap.Pop(pq).(*protocol.Blueprint)
	assert.Equal(t, "svc-1", item.ServiceID) // Lowest (1)
}

func TestFileIntentSource_Priority(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "blueprints-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `[
		{"service_id": "low-prio", "intent": "low", "priority": 1},
		{"service_id": "high-prio", "intent": "high", "priority": 100},
		{"service_id": "med-prio", "intent": "med", "priority": 50}
	]`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	source, err := NewFileIntentSource(tmpFile.Name())
	assert.NoError(t, err)

	ctx := context.Background()

	// 1. High Priority (100)
	bp, err := source.GetNextBlueprint(ctx)
	assert.NoError(t, err)
	if bp == nil {
		t.Fatal("Expected blueprint, got nil")
	}
	assert.Equal(t, "high-prio", bp.ServiceID)

	// 2. Medium Priority (50)
	bp, err = source.GetNextBlueprint(ctx)
	assert.NoError(t, err)
	if bp == nil {
		t.Fatal("Expected blueprint, got nil")
	}
	assert.Equal(t, "med-prio", bp.ServiceID)

	// 3. Low Priority (1)
	bp, err = source.GetNextBlueprint(ctx)
	assert.NoError(t, err)
	if bp == nil {
		t.Fatal("Expected blueprint, got nil")
	}
	assert.Equal(t, "low-prio", bp.ServiceID)

	// 4. End
	bp, err = source.GetNextBlueprint(ctx)
	assert.NoError(t, err)
	assert.Nil(t, bp)
}
