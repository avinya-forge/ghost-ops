package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"ghost-ops/pkg/protocol"
)

// JSONFileStore implements StateStore using a JSON file.
type JSONFileStore struct {
	path string
	mu   sync.RWMutex
}

type storeData struct {
	Services map[string]protocol.ServiceRecord `json:"services"`
	KV       map[string][]byte                 `json:"kv"`
}

// NewJSONFileStore creates a new JSONFileStore.
func NewJSONFileStore(path string) (*JSONFileStore, error) {
	// Check if file exists, if not create empty
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte("{}"), 0600); err != nil {
			return nil, fmt.Errorf("failed to create store file: %w", err)
		}
	}
	return &JSONFileStore{
		path: path,
	}, nil
}

func (s *JSONFileStore) load() (*storeData, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return &storeData{
			Services: make(map[string]protocol.ServiceRecord),
			KV:       make(map[string][]byte),
		}, nil
	}

	// Detect format
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	sd := &storeData{
		Services: make(map[string]protocol.ServiceRecord),
		KV:       make(map[string][]byte),
	}

	_, hasServices := raw["services"]
	_, hasKV := raw["kv"]

	if hasServices || hasKV {
		// New format
		if err := json.Unmarshal(data, sd); err != nil {
			return nil, err
		}
	} else {
		// Old format (map of services) or empty
		var services map[string]protocol.ServiceRecord
		if err := json.Unmarshal(data, &services); err != nil {
			return nil, err
		}
		if services != nil {
			sd.Services = services
		}
	}

	if sd.Services == nil {
		sd.Services = make(map[string]protocol.ServiceRecord)
	}
	if sd.KV == nil {
		sd.KV = make(map[string][]byte)
	}

	return sd, nil
}

// save writes the store atomically: marshal → write to temp file in the same
// directory → fsync → rename over the target. A crash between steps leaves
// either the old file or the new one, never a half-written one (BUG-053).
// The fsync ensures durability across power loss (BUG-057).
func (s *JSONFileStore) save(sd *storeData) error {
	data, err := json.MarshalIndent(sd, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.path)
	tmp, err := os.CreateTemp(dir, ".store-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	// On any error path below, clean up the temp file. Rename consumes it on
	// success, so a stat-miss after Rename is fine.
	defer func() {
		if _, statErr := os.Stat(tmpName); statErr == nil {
			_ = os.Remove(tmpName)
		}
	}()

	if err := os.Chmod(tmpName, 0600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("fsync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("rename temp: %w", err)
	}
	// Best-effort directory fsync so the rename is durable. Failures here
	// don't poison the write — the data is on disk; only the dirent flush is
	// uncertain — so log via the returned error only when truly fatal.
	if d, derr := os.Open(dir); derr == nil {
		_ = d.Sync()
		_ = d.Close()
	}
	return nil
}

// GetService retrieves a service record by ID.
func (s *JSONFileStore) GetService(ctx context.Context, serviceID string) (*protocol.ServiceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sd, err := s.load()
	if err != nil {
		return nil, err
	}

	if sd.Services == nil {
		return nil, nil
	}

	if record, ok := sd.Services[serviceID]; ok {
		return &record, nil
	}
	return nil, nil // Not found
}

// UpdateService updates or creates a service record.
func (s *JSONFileStore) UpdateService(ctx context.Context, record protocol.ServiceRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sd, err := s.load()
	if err != nil {
		return err
	}

	sd.Services[record.ServiceID] = record
	return s.save(sd)
}

// ListServices returns all services.
func (s *JSONFileStore) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sd, err := s.load()
	if err != nil {
		return nil, err
	}

	list := make([]protocol.ServiceRecord, 0, len(sd.Services))
	for _, r := range sd.Services {
		list = append(list, r)
	}
	return list, nil
}

// Get retrieves a value by key.
func (s *JSONFileStore) Get(ctx context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sd, err := s.load()
	if err != nil {
		return nil, err
	}

	val, ok := sd.KV[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return val, nil
}

// Set stores a value by key.
func (s *JSONFileStore) Set(ctx context.Context, key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sd, err := s.load()
	if err != nil {
		return err
	}

	sd.KV[key] = value
	return s.save(sd)
}
