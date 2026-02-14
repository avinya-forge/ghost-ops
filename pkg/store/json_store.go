package store

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"ghost-ops/pkg/protocol"
)

// JSONFileStore implements protocol.StateStore using a local JSON file.
type JSONFileStore struct {
	filePath string
	mu       sync.RWMutex
}

// Ensure JSONFileStore implements protocol.StateStore
var _ protocol.StateStore = (*JSONFileStore)(nil)

// NewJSONFileStore creates a new JSONFileStore with the given file path.
func NewJSONFileStore(filePath string) (*JSONFileStore, error) {
	store := &JSONFileStore{
		filePath: filePath,
	}
	if err := store.ensureFileExists(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *JSONFileStore) ensureFileExists() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := os.Stat(s.filePath)
	if os.IsNotExist(err) {
		// Initialize with empty array
		return os.WriteFile(s.filePath, []byte("[]"), 0644)
	}
	return err
}

func (s *JSONFileStore) load() ([]protocol.ServiceRecord, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}
	var records []protocol.ServiceRecord
	if len(data) == 0 {
		return []protocol.ServiceRecord{}, nil
	}
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *JSONFileStore) save(records []protocol.ServiceRecord) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

// GetService retrieves a service record by ID.
func (s *JSONFileStore) GetService(ctx context.Context, serviceID string) (*protocol.ServiceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records, err := s.load()
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if record.ServiceID == serviceID {
			return &record, nil
		}
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

	found := false
	for i, r := range records {
		if r.ServiceID == record.ServiceID {
			records[i] = record
			found = true
			break
		}
	}

	if !found {
		records = append(records, record)
	}

	return s.save(records)
}

// ListServices returns all services.
func (s *JSONFileStore) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.load()
}
