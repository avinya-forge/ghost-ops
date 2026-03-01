package evolution

import (
	"context"
	"strings"
	"testing"

	"ghost-ops/pkg/llm"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/telemetry"
)

func TestEvaluateCodeGenerationQuality(t *testing.T) {
	// [284] [TEST] Add Evals for Code Generation Quality
	// Since we don't have a real LLM running in CI, we use the MockLLMProvider
	// but add structure for a framework.

	mockLLM := &llm.MockLLMProvider{}
	collector := telemetry.NewInMemoryCollector()
	engine := NewAIEvolutionEngine(mockLLM, collector)

	tests := []struct {
		name      string
		blueprint protocol.Blueprint
		wantValid bool
	}{
		{
			name: "basic service",
			blueprint: protocol.Blueprint{
				ServiceID: "hello-world",
				Intent:    "Write a service that returns hello world",
			},
			wantValid: true,
		},
		{
			name: "calculator service",
			blueprint: protocol.Blueprint{
				ServiceID: "calc-service",
				Intent:    "Write a service that adds two numbers",
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Use the internal LLM generation directly to verify the intermediate text quality
			code, usage, err := engine.llm.GenerateCode(ctx, tt.blueprint)

			if err != nil && tt.wantValid {
				t.Fatalf("Generation failed: %v", err)
			}

			if usage == nil {
				t.Errorf("Expected token usage metrics")
			}

			// Heuristics for valid Go code structure
			if tt.wantValid {
				if !strings.Contains(code, "package main") {
					t.Errorf("Generated code should be a main package")
				}
				if !strings.Contains(code, "func main()") {
					t.Errorf("Generated code should contain a main function")
				}
			}

			// Ensure it compiles
			wasmBytes, err := CompileGo(ctx, code)
			if err != nil && tt.wantValid {
				t.Fatalf("Compilation failed: %v", err)
			}
			if len(wasmBytes) == 0 && tt.wantValid {
				t.Errorf("Generated empty WASM binary")
			}
		})
	}
}
