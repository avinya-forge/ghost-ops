package evolution

import (
	"fmt"
	"math"

	"ghost-ops/pkg/protocol"
)

const (
	defaultCPUMilli = 100
	defaultMemoryMB = 128
)

// ResourceCapacity defines allocatable CPU and memory for scheduling.
type ResourceCapacity struct {
	CPUMilli int
	MemoryMB int
}

// ResourceRequest captures per-service runtime requirements used for bin packing.
type ResourceRequest struct {
	CPUMilli int
	MemoryMB int
}

// ResourceScheduler scores blueprints by resource footprint and enqueues them
// with lower scores first to favor dense placement.
type ResourceScheduler struct {
	queue *PriorityQueue
}

// NewResourceScheduler returns a scheduler backed by PriorityQueue.
func NewResourceScheduler(queue *PriorityQueue) *ResourceScheduler {
	if queue == nil {
		queue = NewPriorityQueue()
	}
	return &ResourceScheduler{queue: queue}
}

// Enqueue estimates a task's packing score and pushes it into the queue.
func (s *ResourceScheduler) Enqueue(bp protocol.Blueprint, capacity ResourceCapacity) (*Task, error) {
	req, err := ParseResourceRequest(bp)
	if err != nil {
		return nil, err
	}
	priority := PackingScore(req, capacity)
	return s.queue.Push(bp, priority), nil
}

// Queue returns the underlying priority queue.
func (s *ResourceScheduler) Queue() *PriorityQueue {
	return s.queue
}

// ParseResourceRequest reads cpu_milli and memory_mb constraints with defaults.
func ParseResourceRequest(bp protocol.Blueprint) (ResourceRequest, error) {
	req := ResourceRequest{CPUMilli: defaultCPUMilli, MemoryMB: defaultMemoryMB}
	if bp.Constraints == nil {
		return req, nil
	}

	if v, ok := bp.Constraints["cpu_milli"]; ok {
		cpu, err := constraintInt(v)
		if err != nil {
			return ResourceRequest{}, fmt.Errorf("cpu_milli must be numeric: %w", err)
		}
		if cpu <= 0 {
			return ResourceRequest{}, fmt.Errorf("cpu_milli must be > 0")
		}
		req.CPUMilli = cpu
	}

	if v, ok := bp.Constraints["memory_mb"]; ok {
		mem, err := constraintInt(v)
		if err != nil {
			return ResourceRequest{}, fmt.Errorf("memory_mb must be numeric: %w", err)
		}
		if mem <= 0 {
			return ResourceRequest{}, fmt.Errorf("memory_mb must be > 0")
		}
		req.MemoryMB = mem
	}

	return req, nil
}

// PackingScore computes a stable integer where lower values are preferred.
// Requests that do not fit in capacity are penalized and deprioritized.
func PackingScore(req ResourceRequest, cap ResourceCapacity) int {
	cpuCap := normalizedCapacity(cap.CPUMilli, defaultCPUMilli)
	memCap := normalizedCapacity(cap.MemoryMB, defaultMemoryMB)

	cpuRatio := float64(req.CPUMilli) / float64(cpuCap)
	memRatio := float64(req.MemoryMB) / float64(memCap)
	maxRatio := math.Max(cpuRatio, memRatio)

	if maxRatio > 1.0 {
		return int(maxRatio*1000) + 10000
	}
	return int(maxRatio * 1000)
}

func normalizedCapacity(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func constraintInt(v interface{}) (int, error) {
	switch n := v.(type) {
	case int:
		return n, nil
	case int32:
		return int(n), nil
	case int64:
		return int(n), nil
	case float64:
		if n != math.Trunc(n) {
			return 0, fmt.Errorf("must be an integer")
		}
		return int(n), nil
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}
