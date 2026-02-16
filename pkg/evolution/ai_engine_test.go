package evolution

import (
	"bytes"
	"context"
	"testing"

	"ghost-ops/pkg/llm"
	"ghost-ops/pkg/protocol"
)

func TestAIEvolutionEngine_Evolve(t *testing.T) {
	// 1. Setup Mock LLM
	// The mock by default returns valid code using guest SDK.
	mockLLM := &llm.MockLLMProvider{}

	// 2. Setup Engine
	engine := NewAIEvolutionEngine(mockLLM)

	// 3. Create Blueprint
	blueprint := protocol.Blueprint{
		ServiceID: "ai-service",
		Intent:    "make a hello world service",
	}

	// 4. Evolve
	ctx := context.Background()
	wasmBytes, err := engine.Evolve(ctx, blueprint)
	if err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	if len(wasmBytes) == 0 {
		t.Fatal("Evolve returned empty bytes")
	}

	// 5. Verify WASM Magic Number
	expectedMagic := []byte{0x00, 0x61, 0x73, 0x6d}
	if !bytes.HasPrefix(wasmBytes, expectedMagic) {
		t.Errorf("Expected WASM magic number, got: %x", wasmBytes[:4])
	}
}
