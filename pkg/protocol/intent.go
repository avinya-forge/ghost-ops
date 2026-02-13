package protocol

import "context"

// Blueprint represents the input for synthesis.
type Blueprint struct {
	ServiceID   string                 `json:"service_id"`
	Intent      string                 `json:"intent"`
	Constraints map[string]interface{} `json:"constraints"`
}

// IntentSource serves as the oracle for system behavior.
type IntentSource interface {
	// GetNextBlueprint returns the next pending blueprint for processing.
	GetNextBlueprint(ctx context.Context) (*Blueprint, error)
}
