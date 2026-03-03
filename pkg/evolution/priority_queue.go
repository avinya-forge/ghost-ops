package evolution

import (
	"container/heap"
	"sync"

	"ghost-ops/pkg/protocol"
)

// Task represents an evolution task with a priority.
type Task struct {
	Blueprint protocol.Blueprint
	Priority  int // Lower number means higher priority (min-heap)
	Index     int // The index of the item in the heap.
}

// PriorityQueue implements a thread-safe min-heap for evolution tasks.
type PriorityQueue struct {
	items taskHeap
	mu    sync.Mutex
}

// NewPriorityQueue creates and initializes a new PriorityQueue.
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		items: make(taskHeap, 0),
	}
	heap.Init(&pq.items)
	return pq
}

// Push adds a new task with the given priority to the queue.
// It returns a pointer to the created Task, which can be used for Update or Remove.
func (pq *PriorityQueue) Push(bp protocol.Blueprint, priority int) *Task {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	task := &Task{
		Blueprint: bp,
		Priority:  priority,
	}
	heap.Push(&pq.items, task)
	return task
}

// Pop removes and returns the task with the highest priority (lowest priority value).
func (pq *PriorityQueue) Pop() (*Task, bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	if len(pq.items) == 0 {
		return nil, false
	}
	item := heap.Pop(&pq.items).(*Task)
	return item, true
}

// Len returns the current number of tasks in the queue.
func (pq *PriorityQueue) Len() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return len(pq.items)
}

// taskHeap is the underlying min-heap implementation.
type taskHeap []*Task

func (h taskHeap) Len() int           { return len(h) }
func (h taskHeap) Less(i, j int) bool { return h[i].Priority < h[j].Priority }
func (h taskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

func (h *taskHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*Task)
	item.Index = n
	*h = append(*h, item)
}

func (h *taskHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.Index = -1 // for safety
	*h = old[0 : n-1]
	return item
}

// Update changes the priority of a specific task in the queue.
func (pq *PriorityQueue) Update(task *Task, bp protocol.Blueprint, priority int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	if task.Index < 0 || task.Index >= len(pq.items) || pq.items[task.Index] != task {
		return
	}
	task.Blueprint = bp
	task.Priority = priority
	heap.Fix(&pq.items, task.Index)
}

// Remove removes a specific task from the queue.
func (pq *PriorityQueue) Remove(task *Task) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	if task.Index < 0 || task.Index >= len(pq.items) || pq.items[task.Index] != task {
		return
	}
	heap.Remove(&pq.items, task.Index)
}
