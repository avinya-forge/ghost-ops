package intent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"ghost-ops/pkg/protocol"
)

// FileIntentSource reads blueprints from a JSON file.
type FileIntentSource struct {
	blueprints   []protocol.Blueprint
	mu           sync.Mutex
	index        int
	filePath     string
	lastModified time.Time
}

// NewFileIntentSource creates a new FileIntentSource and loads blueprints from the given file.
func NewFileIntentSource(filepath string) (*FileIntentSource, error) {
	f := &FileIntentSource{
		filePath: filepath,
		index:    0,
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

	f.blueprints = blueprints
	f.lastModified = info.ModTime()
	return true, nil
}

// GetNextBlueprint returns the next pending blueprint from the loaded list.
// It returns nil, nil when all blueprints have been consumed.
func (f *FileIntentSource) GetNextBlueprint(ctx context.Context) (*protocol.Blueprint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.index >= len(f.blueprints) {
		reloaded, err := f.load()
		if err != nil {
			return nil, err
		}
		if reloaded {
			f.index = 0
		}
	}

	if f.index >= len(f.blueprints) {
		return nil, nil // End of list
	}

	bp := f.blueprints[f.index]
	f.index++
	return &bp, nil
}
