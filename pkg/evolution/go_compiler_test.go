package evolution

import (
	"bytes"
	"context"
	"os"
	"testing"

	"ghost-ops/pkg/protocol"
)

func TestGoCompilerEngine_Evolve(t *testing.T) {
	engine := NewGoCompilerEngine()

	// Minimal Go source code
	sourceCode := `package main
import "fmt"
func main() {
	fmt.Println("Hello from Test")
}`

	blueprint := protocol.Blueprint{
		ServiceID: "test-service",
		Constraints: map[string]interface{}{
			"source_code": sourceCode,
		},
	}

	ctx := context.Background()
	wasmBytes, err := engine.Evolve(ctx, blueprint)
	if err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	if len(wasmBytes) == 0 {
		t.Fatal("Evolve returned empty bytes")
	}

	// WASM Magic Number: \0asm
	expectedMagic := []byte{0x00, 0x61, 0x73, 0x6d}
	if !bytes.HasPrefix(wasmBytes, expectedMagic) {
		t.Errorf("Expected WASM magic number, got: %x", wasmBytes[:4])
	}
}

func TestGoCompilerEngine_SecurityScanFails(t *testing.T) {
	engine := NewGoCompilerEngine()

	// Malicious Go source code that should trigger a gosec finding (G401/G501: MD5)
	sourceCode := `package main
import (
	"crypto/md5"
	"fmt"
)
func main() {
	h := md5.New()
	fmt.Println(h)
}`

	blueprint := protocol.Blueprint{
		ServiceID: "malicious-service",
		Constraints: map[string]interface{}{
			"source_code": sourceCode,
		},
	}

	ctx := context.Background()
	_, err := engine.Evolve(ctx, blueprint)
	if err == nil {
		t.Fatal("Evolve succeeded but should have failed due to gosec security scan")
	}
}

func TestCompileGo_MissingGosec(t *testing.T) {
	// Temporarily remove gosec from PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", "")

	sourceCode := `package main
import "fmt"
func main() {
	fmt.Println("Hello")
}`

	_, err := CompileGo(context.Background(), sourceCode)
	if err == nil {
		t.Fatal("Expected error when gosec is not in PATH, got nil")
	}

	if err.Error() != "gosec not found in PATH" {
		t.Fatalf("Expected 'gosec not found in PATH', got '%v'", err)
	}
}
