package protocol

import (
	"context"
	"errors"
	"sync"
)

// InMemoryStateStore is a thread-safe in-memory implementation of StateStore.
type InMemoryStateStore struct {
	mu       sync.RWMutex
	services map[string]ServiceRecord
}

// NewInMemoryStateStore creates a new instance.
func NewInMemoryStateStore() *InMemoryStateStore {
	return &InMemoryStateStore{
		services: make(map[string]ServiceRecord),
	}
}

// GetService retrieves a service record by ID.
func (s *InMemoryStateStore) GetService(ctx context.Context, serviceID string) (*ServiceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.services[serviceID]
	if !ok {
		return nil, errors.New("service not found")
	}
	return &record, nil
}

// UpdateService updates or creates a service record.
func (s *InMemoryStateStore) UpdateService(ctx context.Context, record ServiceRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.services[record.ServiceID] = record
	return nil
}

// ListServices returns all services.
func (s *InMemoryStateStore) ListServices(ctx context.Context) ([]ServiceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var records []ServiceRecord
	for _, record := range s.services {
		records = append(records, record)
	}
	return records, nil
}
