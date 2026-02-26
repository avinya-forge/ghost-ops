package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
)

var serviceIDRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// Blueprint represents the input for synthesis.
type Blueprint struct {
	ServiceID    string                 `json:"service_id"`
	Intent       string                 `json:"intent"`
	Constraints  map[string]interface{} `json:"constraints"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Priority     int                    `json:"priority,omitempty"` // Higher value = Higher priority
}

// Validate checks if the blueprint is valid.
func (b *Blueprint) Validate() error {
	if b.ServiceID == "" {
		return fmt.Errorf("service_id is required")
	}
	if !serviceIDRegex.MatchString(b.ServiceID) {
		return fmt.Errorf("service_id must contain only alphanumeric characters and hyphens")
	}
	if b.Intent == "" {
		return fmt.Errorf("intent is required")
	}
	if len(b.Intent) > 10000 { // Arbitrary limit for now
		return fmt.Errorf("intent is too long (max 10000 chars)")
	}

	// Validate Dependencies
	for _, dep := range b.Dependencies {
		if !serviceIDRegex.MatchString(dep) {
			return fmt.Errorf("dependency service_id '%s' must contain only alphanumeric characters and hyphens", dep)
		}
		if dep == b.ServiceID {
			return fmt.Errorf("service cannot depend on itself")
		}
	}

	// Check constraints size
	if b.Constraints != nil {
		bytes, err := json.Marshal(b.Constraints)
		if err != nil {
			return fmt.Errorf("failed to marshal constraints for validation: %w", err)
		}
		if len(bytes) > 1024*1024 { // 1MB
			return fmt.Errorf("constraints size exceeds 1MB limit")
		}
	}
	return nil
}

// IntentSource serves as the oracle for system behavior.
type IntentSource interface {
	// GetNextBlueprint returns the next pending blueprint for processing.
	GetNextBlueprint(ctx context.Context) (*Blueprint, error)
}
