package evolution

import (
	"bytes"
	"context"
	"testing"

	"ghost-ops/pkg/protocol"
)

func TestMockEvolutionEngine(t *testing.T) {
	// Test default minimal WASM
	engine := NewMockEvolutionEngine("")
	bp := protocol.Blueprint{Intent: "test"}
	wasm, err := engine.Evolve(context.Background(), bp)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	if !bytes.Equal(wasm, expected) {
		t.Errorf("Expected %x, got %x", expected, wasm)
	}
}
