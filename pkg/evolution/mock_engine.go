package evolution

import (
	"context"
	"fmt"
	"os"

	"ghost-ops/pkg/protocol"
)

// MockEvolutionEngine generates a fixed WASM binary for any blueprint.
type MockEvolutionEngine struct {
	wasmPath string
}

// NewMockEvolutionEngine creates a new MockEvolutionEngine.
// If wasmPath is empty, it returns a minimal valid empty WASM module.
func NewMockEvolutionEngine(wasmPath string) *MockEvolutionEngine {
	return &MockEvolutionEngine{
		wasmPath: wasmPath,
	}
}

// Evolve returns the WASM binary.
func (e *MockEvolutionEngine) Evolve(ctx context.Context, blueprint protocol.Blueprint) ([]byte, error) {
	if e.wasmPath == "" {
		// Minimal valid WASM module: magic number + version 1
		return []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}, nil
	}

	data, err := os.ReadFile(e.wasmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read WASM file: %w", err)
	}
	return data, nil
}
