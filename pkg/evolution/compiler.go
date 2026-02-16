package evolution

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CompileGo compiles Go source code into WASM.
// It creates a temporary directory, writes the source code to a file,
// and runs `go build` with GOOS=wasip1 GOARCH=wasm.
func CompileGo(ctx context.Context, sourceCode string) ([]byte, error) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "ghost-ops-build-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(inputFile, []byte(sourceCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write source code: %w", err)
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
