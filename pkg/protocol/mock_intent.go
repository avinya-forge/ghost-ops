package protocol

import "context"

// MockIntentSource is a mock implementation of IntentSource.
type MockIntentSource struct {
	Blueprints []Blueprint
	index      int
}

// GetNextBlueprint returns the next blueprint in the sequence.
func (m *MockIntentSource) GetNextBlueprint(ctx context.Context) (*Blueprint, error) {
	if m.index >= len(m.Blueprints) {
		return nil, nil
	}
	bp := m.Blueprints[m.index]
	m.index++
	return &bp, nil
}
