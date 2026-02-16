package evolution

import (
	"context"
	"fmt"

	"ghost-ops/pkg/protocol"
)

// AIEvolutionEngine uses an LLM to generate code from intent, then compiles it.
type AIEvolutionEngine struct {
	llm protocol.LLMProvider
}

// NewAIEvolutionEngine creates a new AIEvolutionEngine with the given LLM provider.
func NewAIEvolutionEngine(llm protocol.LLMProvider) *AIEvolutionEngine {
	return &AIEvolutionEngine{
		llm: llm,
	}
}

// Evolve generates a WASM binary from the given blueprint using AI.
func (e *AIEvolutionEngine) Evolve(ctx context.Context, blueprint protocol.Blueprint) ([]byte, error) {
	// 1. Generate Source Code from Intent
	sourceCode, err := e.llm.GenerateCode(ctx, blueprint.Intent)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code from intent: %w", err)
	}

	// 2. Compile Source Code
	wasmBytes, err := CompileGo(ctx, sourceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to compile generated code: %w", err)
	}

	return wasmBytes, nil
}
