package runtime

import "ghost-ops/pkg/config"
import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/telemetry"
)

func TestWazeroRuntimeHost_AsyncLoad(t *testing.T) {
	ctx := context.Background()
	store := protocol.NewInMemoryStateStore()
	collector := telemetry.NewInMemoryCollector()
	host, err := NewWazeroRuntimeHost(ctx, store, collector, config.CapabilitiesConfig{})
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// Use the Go Compiler to build the kv-service example
	// which enters an infinite loop in main() -> guest.Start()
	engine := evolution.NewGoCompilerEngine()
	// Resolves to repo root/examples/services/kv-service/main.go
	srcPath, err := filepath.Abs("../../examples/services/kv-service/main.go")
	if err != nil {
		t.Fatalf("Failed to resolve source path: %v", err)
	}

	bp := protocol.Blueprint{Constraints: map[string]interface{}{"source_path": srcPath}}
	wasmBytes, err := engine.Evolve(ctx, bp)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	serviceID := "async-test-service"
	version := "v1"

	start := time.Now()
	// This should return IMMEDIATELY because InstantiateModule is async.
	// The module runs an infinite loop, so if it were sync, LoadModule would block forever.
	if err := host.LoadModule(ctx, serviceID, version, wasmBytes); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}
	duration := time.Since(start)

	// Assert it returned quickly. Allow generous margin for Wazero JIT compilation,
	// which is significantly slower under -race. A hung LoadModule (sync regression)
	// would block indefinitely because the module runs an infinite loop.
	if duration > 30*time.Second {
		t.Errorf("LoadModule took too long: %v (expected < 30s)", duration)
	} else {
		t.Logf("LoadModule returned in %v", duration)
	}

	// Set active version to allow Invoke
	if err := host.SetActiveVersion(ctx, serviceID, version); err != nil {
		t.Fatalf("Failed to set active version: %v", err)
	}

	// Verify the module eventually becomes responsive
	// Invoke waits for the module to be ready (call next_command)
	// It has a 5s timeout, which should be enough for instantiation
	_, err = host.Invoke(ctx, serviceID, "Handle", []byte("hello"))
	if err != nil {
		t.Fatalf("Failed to invoke module: %v", err)
	}
}
