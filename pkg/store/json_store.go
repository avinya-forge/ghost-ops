package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"ghost-ops/pkg/protocol"
)

// JSONFileStore implements StateStore using a JSON file.
type JSONFileStore struct {
	path string
	mu   sync.RWMutex
}

// NewJSONFileStore creates a new JSONFileStore.
func NewJSONFileStore(path string) (*JSONFileStore, error) {
	// Check if file exists, if not create empty
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			return nil, fmt.Errorf("failed to create store file: %w", err)
		}
	}
	return &JSONFileStore{
		path: path,
	}, nil
}

func (s *JSONFileStore) load() (map[string]protocol.ServiceRecord, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return make(map[string]protocol.ServiceRecord), nil
	}
	var records map[string]protocol.ServiceRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *JSONFileStore) save(records map[string]protocol.ServiceRecord) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// GetService retrieves a service record by ID.
func (s *JSONFileStore) GetService(ctx context.Context, serviceID string) (*protocol.ServiceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records, err := s.load()
	if err != nil {
		return nil, err
	}

	if record, ok := records[serviceID]; ok {
		return &record, nil
	}
	return nil, nil // Not found
}

// UpdateService updates or creates a service record.
func (s *JSONFileStore) UpdateService(ctx context.Context, record protocol.ServiceRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	records, err := s.load()
	if err != nil {
		return err
	}

	records[record.ServiceID] = record
	return s.save(records)
}

// ListServices returns all services.
func (s *JSONFileStore) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records, err := s.load()
	if err != nil {
		return nil, err
	}

	list := make([]protocol.ServiceRecord, 0, len(records))
	for _, r := range records {
		list = append(list, r)
	}
	return list, nil
}
