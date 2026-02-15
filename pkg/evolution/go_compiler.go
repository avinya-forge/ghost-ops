package evolution

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "ghost-ops-build-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	var inputFile string
	if hasSource {
		inputFile = filepath.Join(tmpDir, "main.go")
		if err := os.WriteFile(inputFile, []byte(sourceCode), 0644); err != nil {
			return nil, fmt.Errorf("failed to write source code: %w", err)
		}
	} else {
		// Verify source path exists
		if _, err := os.Stat(sourcePath); err != nil {
			return nil, fmt.Errorf("failed to access source path: %w", err)
		}
		inputFile = sourcePath
	}

	outputFile := filepath.Join(tmpDir, "output.wasm")

	// Run go build
	cmd := exec.CommandContext(ctx, "go", "build", "-o", outputFile, inputFile)
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")

	// Capture output for debugging
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("compilation failed: %v, output: %s", err, string(output))
	}

	// Read output WASM
	wasmBytes, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output WASM: %w", err)
	}

	return wasmBytes, nil
}
