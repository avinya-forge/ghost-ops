package evolution

import (
	"context"
	"fmt"
	"os"

	"ghost-ops/pkg/protocol"
)

// GoCompilerEngine compiles Go source code into WASM.
type GoCompilerEngine struct{}

// NewGoCompilerEngine creates a new GoCompilerEngine.
func NewGoCompilerEngine() *GoCompilerEngine {
	return &GoCompilerEngine{}
}

// Evolve generates a WASM binary from the given blueprint.
// It expects "source_code" (string) or "source_path" (string) in constraints.
func (e *GoCompilerEngine) Evolve(ctx context.Context, blueprint protocol.Blueprint) ([]byte, error) {
	// Check for source code or path in constraints
	sourceCode, hasSource := blueprint.Constraints["source_code"].(string)
	sourcePath, hasPath := blueprint.Constraints["source_path"].(string)

	if !hasSource && !hasPath {
		return nil, fmt.Errorf("blueprint constraints must contain 'source_code' or 'source_path'")
	}

	if hasSource {
		return CompileGo(ctx, sourceCode)
	}

	// Handle source_path case
	// Read the file content and use CompileGo
	codeBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read source path: %w", err)
	}

	return CompileGo(ctx, string(codeBytes))
}
