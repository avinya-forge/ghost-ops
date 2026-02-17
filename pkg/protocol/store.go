package protocol

import (
	"context"
	"time"
)

// ServiceState represents the lifecycle state of a service.
type ServiceState string

const (
	StateActive  ServiceState = "ACTIVE"
	StateShadow  ServiceState = "SHADOW"
	StatePurging ServiceState = "PURGING"
)

// ServiceRecord represents a row in the Services table.
type ServiceRecord struct {
	ServiceID          string       `json:"service_id"`
	Version            int          `json:"version"`
	WASMHash           string       `json:"wasm_hash"`
	CurrentState       ServiceState `json:"current_state"`
	SynthesisTimestamp time.Time    `json:"synthesis_timestamp"`
}

// StateStore defines the interface for the global truth registry.
type StateStore interface {
	// GetService retrieves a service record by ID.
	GetService(ctx context.Context, serviceID string) (*ServiceRecord, error)

	// UpdateService updates or creates a service record.
	UpdateService(ctx context.Context, record ServiceRecord) error

	// ListServices returns all services.
	ListServices(ctx context.Context) ([]ServiceRecord, error)

	// Get retrieves a value by key.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value by key.
	Set(ctx context.Context, key string, value []byte) error
}
