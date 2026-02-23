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
	kv       map[string][]byte
}

// NewInMemoryStateStore creates a new instance.
func NewInMemoryStateStore() *InMemoryStateStore {
	return &InMemoryStateStore{
		services: make(map[string]ServiceRecord),
		kv:       make(map[string][]byte),
	}
}

// GetService retrieves a service record by ID.
func (s *InMemoryStateStore) GetService(ctx context.Context, serviceID string) (*ServiceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.services[serviceID]
	if !ok {
		return nil, nil // Return nil if not found
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

// Get retrieves a value by key.
func (s *InMemoryStateStore) Get(ctx context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.kv[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return val, nil
}

// Set stores a value by key.
func (s *InMemoryStateStore) Set(ctx context.Context, key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.kv[key] = value
	return nil
}
