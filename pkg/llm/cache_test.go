package llm

import (
	"context"
	"testing"

	"ghost-ops/pkg/protocol"
)

// MockProvider for testing
type MockProvider struct {
	calls map[string]int
}

func (m *MockProvider) GenerateCode(ctx context.Context, blueprint protocol.Blueprint) (string, *protocol.TokenUsage, error) {
	if m.calls == nil {
		m.calls = make(map[string]int)
	}
	m.calls[blueprint.Intent]++
	return "mock code", &protocol.TokenUsage{InputTokens: 10, OutputTokens: 20, TotalTokens: 30}, nil
}

func TestCachedLLMProvider_GenerateCode(t *testing.T) {
	mock := &MockProvider{}
	cache := NewInMemoryCache()
	provider := NewCachedLLMProvider(mock, cache)

	ctx := context.Background()
	blueprint := protocol.Blueprint{
		ServiceID: "test-service",
		Intent:    "Test Intent",
		Constraints: map[string]interface{}{
			"language": "go",
		},
	}

	// First call - should hit provider
	code, usage, err := provider.GenerateCode(ctx, blueprint)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	if code != "mock code" {
		t.Errorf("Expected 'mock code', got '%s'", code)
	}
	if usage.TotalTokens != 30 {
		t.Errorf("Expected 30 tokens, got %d", usage.TotalTokens)
	}
	if mock.calls["Test Intent"] != 1 {
		t.Errorf("Expected 1 call to provider, got %d", mock.calls["Test Intent"])
	}

	// Second call - should hit cache
	code, usage, err = provider.GenerateCode(ctx, blueprint)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
	if code != "mock code" {
		t.Errorf("Expected 'mock code', got '%s'", code)
	}
	if usage.TotalTokens != 30 {
		t.Errorf("Expected 30 tokens, got %d", usage.TotalTokens)
	}
	if mock.calls["Test Intent"] != 1 {
		t.Errorf("Expected still 1 call to provider, got %d", mock.calls["Test Intent"])
	}

	// Different blueprint - should hit provider again
	blueprint2 := protocol.Blueprint{
		ServiceID: "test-service-2",
		Intent:    "Different Intent",
		Constraints: map[string]interface{}{
			"language": "rust",
		},
	}

	_, _, err = provider.GenerateCode(ctx, blueprint2)
	if err != nil {
		t.Fatalf("Third call failed: %v", err)
	}
	if mock.calls["Different Intent"] != 1 {
		t.Errorf("Expected 1 call for new intent, got %d", mock.calls["Different Intent"])
	}
}
