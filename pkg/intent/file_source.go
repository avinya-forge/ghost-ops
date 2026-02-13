package intent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"ghost-ops/pkg/protocol"
)

// FileIntentSource reads blueprints from a JSON file.
type FileIntentSource struct {
	blueprints []protocol.Blueprint
	mu         sync.Mutex
	index      int
}

// NewFileIntentSource creates a new FileIntentSource and loads blueprints from the given file.
func NewFileIntentSource(filepath string) (*FileIntentSource, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var blueprints []protocol.Blueprint
	if len(data) > 0 {
		if err := json.Unmarshal(data, &blueprints); err != nil {
			return nil, fmt.Errorf("failed to unmarshal blueprints: %w", err)
		}
	}

	return &FileIntentSource{
		blueprints: blueprints,
		index:      0,
	}, nil
}

// GetNextBlueprint returns the next pending blueprint from the loaded list.
// It returns nil, nil when all blueprints have been consumed.
func (f *FileIntentSource) GetNextBlueprint(ctx context.Context) (*protocol.Blueprint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.index >= len(f.blueprints) {
		return nil, nil // End of list
	}

	bp := f.blueprints[f.index]
	f.index++
	return &bp, nil
}
