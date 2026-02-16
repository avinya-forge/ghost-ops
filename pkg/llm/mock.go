package llm

import (
	"context"
)

// MockLLMProvider is a mock implementation of LLMProvider.
type MockLLMProvider struct {
	// Optional: Fixed response to return.
	// If empty, returns a default "Hello World" service.
	Response string
}

// GenerateCode returns a hardcoded Go service code.
func (m *MockLLMProvider) GenerateCode(ctx context.Context, intent string) (string, error) {
	if m.Response != "" {
		return m.Response, nil
	}

	// Return a simple standalone WASM module that doesn't require external dependencies
	// to ensure tests run reliably in all environments.
	return `package main

import "fmt"

func main() {
	fmt.Println("Hello from AI-generated code")
}
`, nil
}
