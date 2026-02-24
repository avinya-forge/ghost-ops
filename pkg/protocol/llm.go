package protocol

import "context"

// TokenUsage tracks the token usage of an LLM request.
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// LLMProvider defines the interface for Large Language Model interactions.
type LLMProvider interface {
	// GenerateCode transforms a natural language intent into valid source code using the provided blueprint.
	GenerateCode(ctx context.Context, blueprint Blueprint) (string, *TokenUsage, error)
}
