package runtime

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/protocol"
)

// TestWazeroRuntimeHost_Integration verifies the runtime host using the Guest SDK.
// The raw WASM tests were removed because hand-writing a Reactor module (looping next_command)
// in raw bytes is too complex and error-prone.
func TestWazeroRuntimeHost_Integration(t *testing.T) {
	ctx := context.Background()
	store := protocol.NewInMemoryStateStore()
	host, err := NewWazeroRuntimeHost(ctx, store)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// Use the Go Compiler to build the kv-service example for testing
	engine := evolution.NewGoCompilerEngine()

	// Resolves to repo root/examples/services/kv-service/main.go
	srcPath, err := filepath.Abs("../../examples/services/kv-service/main.go")
	if err != nil {
		t.Fatalf("Failed to resolve source path: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(srcPath); err != nil {
		t.Fatalf("Source file not found at %s: %v", srcPath, err)
	}

	bp := protocol.Blueprint{
		Constraints: map[string]interface{}{
			"source_path": srcPath,
		},
	}

	wasmBytes, err := engine.Evolve(ctx, bp)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	serviceID := "integration-test-service"

	// Test LoadModule
	if err := host.LoadModule(ctx, serviceID, wasmBytes); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}
	defer host.UnloadModule(ctx, serviceID)

	// Set initial value in store
	if err := store.Set(ctx, "hello", []byte("world")); err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Test Invoke "Handle"
	// The kv-service example implements Handle which calls Get(payload)
	payload := []byte("hello")
	output, err := host.Invoke(ctx, serviceID, "Handle", payload)
	if err != nil {
		t.Fatalf("Failed to invoke method: %v", err)
	}

	expected := "value: world"
	if string(output) != expected {
		t.Errorf("Expected output %q, got %q", expected, string(output))
	}

	// Test UnloadModule
	if err := host.UnloadModule(ctx, serviceID); err != nil {
		t.Fatalf("Failed to unload module: %v", err)
	}

	// Test Invoke after unload (should fail)
	if _, err := host.Invoke(ctx, serviceID, "Handle", payload); err == nil {
		t.Error("Expected error after unload, got nil")
	}
}
