package runtime

import "ghost-ops/pkg/config"
import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/telemetry"
)

// TestWazeroRuntimeHost_Integration verifies the runtime host using the Guest SDK.
// The raw WASM tests were removed because hand-writing a Reactor module (looping next_command)
// in raw bytes is too complex and error-prone.
func TestWazeroRuntimeHost_Integration(t *testing.T) {
	ctx := context.Background()
	store := protocol.NewInMemoryStateStore()
	collector := telemetry.NewInMemoryCollector()
	host, err := NewWazeroRuntimeHost(ctx, store, collector, config.CapabilitiesConfig{})
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// Use the Go Compiler to build the kv-service example for testing
	engine := evolution.NewGoCompilerEngine()

	// Resolves to repo root/examples/services/kv-service/main.go
	// We need to find the repo root. Assuming running from pkg/runtime
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
	version := "v1"

	// Test LoadModule
	if err := host.LoadModule(ctx, serviceID, version, wasmBytes); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Test SetActiveVersion
	if err := host.SetActiveVersion(ctx, serviceID, version); err != nil {
		t.Fatalf("Failed to set active version: %v", err)
	}

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

	// Test UnloadVersion
	if err := host.UnloadVersion(ctx, serviceID, version); err != nil {
		t.Fatalf("Failed to unload module: %v", err)
	}

	// Test Invoke after unload (should fail because active version unloaded)
	if _, err := host.Invoke(ctx, serviceID, "Handle", payload); err == nil {
		t.Error("Expected error after unload, got nil")
	}
}

func TestWazeroRuntimeHost_Versioning(t *testing.T) {
	ctx := context.Background()
	store := protocol.NewInMemoryStateStore()
	collector := telemetry.NewInMemoryCollector()
	host, err := NewWazeroRuntimeHost(ctx, store, collector, config.CapabilitiesConfig{})
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	engine := evolution.NewGoCompilerEngine()
	srcPath, _ := filepath.Abs("../../examples/services/kv-service/main.go")
	bp := protocol.Blueprint{Constraints: map[string]interface{}{"source_path": srcPath}}
	wasmBytes, err := engine.Evolve(ctx, bp)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	serviceID := "versioning-test-service"

	// Load v1
	if err := host.LoadModule(ctx, serviceID, "v1", wasmBytes); err != nil {
		t.Fatalf("Failed to load v1: %v", err)
	}
	if err := host.SetActiveVersion(ctx, serviceID, "v1"); err != nil {
		t.Fatalf("Failed to activate v1: %v", err)
	}

	// Invoke v1
	if _, err := host.Invoke(ctx, serviceID, "Handle", []byte("key")); err != nil {
		t.Fatalf("Failed to invoke v1: %v", err)
	}

	// Load v2 (same binary)
	if err := host.LoadModule(ctx, serviceID, "v2", wasmBytes); err != nil {
		t.Fatalf("Failed to load v2: %v", err)
	}

	// Switch to v2
	if err := host.SetActiveVersion(ctx, serviceID, "v2"); err != nil {
		t.Fatalf("Failed to activate v2: %v", err)
	}

	// Unload v1
	if err := host.UnloadVersion(ctx, serviceID, "v1"); err != nil {
		t.Fatalf("Failed to unload v1: %v", err)
	}

	// Invoke v2 should work
	if _, err := host.Invoke(ctx, serviceID, "Handle", []byte("key")); err != nil {
		t.Fatalf("Failed to invoke v2 after unloading v1: %v", err)
	}

	// Unload v2
	host.UnloadVersion(ctx, serviceID, "v2")

	// Invoke should fail
	if _, err := host.Invoke(ctx, serviceID, "Handle", []byte("key")); err == nil {
		t.Fatal("Expected error after unloading all versions")
	}
}
