package llm

import (
	"context"

	"ghost-ops/pkg/protocol"
)

// MockLLMProvider is a mock implementation of LLMProvider.
type MockLLMProvider struct {
	// Optional: Fixed response to return.
	// If empty, returns a default "Hello World" service.
	Response string
}

// GenerateCode returns a hardcoded Go service code.
func (m *MockLLMProvider) GenerateCode(ctx context.Context, blueprint protocol.Blueprint) (string, *protocol.TokenUsage, error) {
	if m.Response != "" {
		return m.Response, &protocol.TokenUsage{InputTokens: 10, OutputTokens: 10, TotalTokens: 20}, nil
	}

	// Return a simple standalone WASM module that doesn't require external dependencies
	// to ensure tests run reliably in all environments.
	return `package main

import "fmt"

func main() {
	fmt.Println("Hello from AI-generated code")
}
`, &protocol.TokenUsage{InputTokens: 10, OutputTokens: 10, TotalTokens: 20}, nil
}
