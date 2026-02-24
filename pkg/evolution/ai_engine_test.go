package evolution

import (
	"bytes"
	"context"
	"testing"

	"ghost-ops/pkg/llm"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/telemetry"
)

func TestAIEvolutionEngine_Evolve(t *testing.T) {
	// 1. Setup Mock LLM
	// The mock by default returns valid code using guest SDK.
	mockLLM := &llm.MockLLMProvider{}

	// 2. Setup Collector
	collector := telemetry.NewInMemoryCollector()

	// 3. Setup Engine
	engine := NewAIEvolutionEngine(mockLLM, collector)

	// 4. Create Blueprint
	blueprint := protocol.Blueprint{
		ServiceID: "ai-service",
		Intent:    "make a hello world service",
	}

	// 5. Evolve
	ctx := context.Background()
	wasmBytes, err := engine.Evolve(ctx, blueprint)
	if err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	if len(wasmBytes) == 0 {
		t.Fatal("Evolve returned empty bytes")
	}

	// 6. Verify WASM Magic Number
	expectedMagic := []byte{0x00, 0x61, 0x73, 0x6d}
	if !bytes.HasPrefix(wasmBytes, expectedMagic) {
		t.Errorf("Expected WASM magic number, got: %x", wasmBytes[:4])
	}

	// 7. Verify Metrics
	// The InMemoryCollector uses "name{k=v,k=v}" format for keys, sorted by label key.
	// Expected key: llm_tokens_total{service_id=ai-service,type=total}
	// MockLLM returns 20 total tokens.
	snapshot := collector.Snapshot()

	expectedKey := "llm_tokens_total{service_id=ai-service,type=total}"
	val, ok := snapshot[expectedKey]
	if !ok {
		t.Errorf("Metric %s not found in snapshot", expectedKey)
	} else {
		count, ok := val.(int64)
		if !ok {
			t.Errorf("Metric %s is not int64", expectedKey)
		} else if count != 20 {
			t.Errorf("Expected 20 tokens, got %d", count)
		}
	}
}
