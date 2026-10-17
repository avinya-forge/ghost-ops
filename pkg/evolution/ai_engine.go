package evolution

import (
	"context"
	"fmt"
	"strings"

	"ghost-ops/pkg/protocol"
)

// AIEvolutionEngine uses an LLM to generate code from intent, then compiles it.
type AIEvolutionEngine struct {
	llm       protocol.LLMProvider
	collector protocol.MetricsCollector
}

// NewAIEvolutionEngine creates a new AIEvolutionEngine with the given LLM provider and collector.
func NewAIEvolutionEngine(llm protocol.LLMProvider, collector protocol.MetricsCollector) *AIEvolutionEngine {
	return &AIEvolutionEngine{
		llm:       llm,
		collector: collector,
	}
}

// Evolve generates a WASM binary from the given blueprint using AI.
func (e *AIEvolutionEngine) Evolve(ctx context.Context, blueprint protocol.Blueprint) ([]byte, error) {
	// Sanitize the blueprint before processing
	blueprint.Sanitize()

	// 1. Generate Source Code from Blueprint
	sourceCode, usage, err := e.llm.GenerateCode(ctx, blueprint)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code from intent: %w", err)
	}

	// Inject the // ghost-ops:generated tag at the very beginning of the source code if not already present.
	if !strings.HasPrefix(sourceCode, "// ghost-ops:generated") {
		sourceCode = "// ghost-ops:generated\n" + sourceCode
	}

	// Record token usage metrics
	// This helps in tracking the cost and volume of LLM usage per service.
	if usage != nil && e.collector != nil {
		e.collector.Counter("llm_tokens_total", int64(usage.InputTokens), map[string]string{"type": "input", "service_id": blueprint.ServiceID})
		e.collector.Counter("llm_tokens_total", int64(usage.OutputTokens), map[string]string{"type": "output", "service_id": blueprint.ServiceID})
		e.collector.Counter("llm_tokens_total", int64(usage.TotalTokens), map[string]string{"type": "total", "service_id": blueprint.ServiceID})
	}

	// Inject required // ghost-ops:generated tag (EU AI Act Art 13)
	sourceCode = "// ghost-ops:generated\n\n" + sourceCode

	// 2. Compile Source Code
	wasmBytes, err := CompileGo(ctx, sourceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to compile generated code: %w", err)
	}

	return wasmBytes, nil
}
