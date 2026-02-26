package intent

import (
	"container/heap"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"ghost-ops/pkg/protocol"
)

// PriorityQueue implements heap.Interface and holds Blueprints.
type PriorityQueue []*protocol.Blueprint

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest priority so we use Greater than.
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x any) {
	item := x.(*protocol.Blueprint)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

// FileIntentSource reads blueprints from a JSON file.
type FileIntentSource struct {
	pq           PriorityQueue
	mu           sync.Mutex
	filePath     string
	lastModified time.Time
}

// NewFileIntentSource creates a new FileIntentSource and loads blueprints from the given file.
func NewFileIntentSource(filepath string) (*FileIntentSource, error) {
	f := &FileIntentSource{
		filePath: filepath,
		pq:       make(PriorityQueue, 0),
	}

	if _, err := f.load(); err != nil {
		return nil, err
	}

	return f, nil
}

// load reads the file and updates the blueprints if changed.
// Returns true if blueprints were reloaded.
func (f *FileIntentSource) load() (bool, error) {
	info, err := os.Stat(f.filePath)
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	// If not first load and not modified
	if !f.lastModified.IsZero() && !info.ModTime().After(f.lastModified) {
		return false, nil
	}

	data, err := os.ReadFile(f.filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	var blueprints []protocol.Blueprint
	if len(data) > 0 {
		if err := json.Unmarshal(data, &blueprints); err != nil {
			return false, fmt.Errorf("failed to unmarshal blueprints: %w", err)
		}
	}

	// Rebuild Priority Queue
	f.pq = make(PriorityQueue, 0, len(blueprints))
	for i := range blueprints {
		// Store pointers to elements
		bp := blueprints[i]
		heap.Push(&f.pq, &bp)
	}

	f.lastModified = info.ModTime()
	return true, nil
}

// GetNextBlueprint returns the next pending blueprint for processing.
// It returns nil, nil when all blueprints have been consumed.
func (f *FileIntentSource) GetNextBlueprint(ctx context.Context) (*protocol.Blueprint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.pq.Len() == 0 {
		reloaded, err := f.load()
		if err != nil {
			return nil, err
		}
		if !reloaded {
			return nil, nil // End of list and no updates
		}
	}

	if f.pq.Len() == 0 {
		return nil, nil // Still empty (e.g. empty file)
	}

	item := heap.Pop(&f.pq).(*protocol.Blueprint)
	return item, nil
}
