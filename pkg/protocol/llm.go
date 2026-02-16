package protocol

import "context"

// LLMProvider defines the interface for Large Language Model interactions.
type LLMProvider interface {
	// GenerateCode transforms a natural language intent into valid source code.
	GenerateCode(ctx context.Context, intent string) (string, error)
}
