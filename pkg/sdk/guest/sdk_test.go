package guest_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/runtime"
	"ghost-ops/pkg/store"
	"ghost-ops/pkg/telemetry"
)

func TestGuestSDK(t *testing.T) {
	ctx := context.Background()

	// 1. Compile Service
	engine := evolution.NewGoCompilerEngine()
	// Resolves to repo root/examples/services/kv-service/main.go
	// Since test is in pkg/sdk/guest, we go up 3 levels.
	srcPath, err := filepath.Abs("../../../examples/services/kv-service/main.go")
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

	// 2. Setup Host
	storePath := "test_store.json"
	st, err := store.NewJSONFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer os.Remove(storePath)

	// Set initial value
	if err := st.Set(ctx, "foo", []byte("bar")); err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	collector := telemetry.NewInMemoryCollector()
	host, err := runtime.NewWazeroRuntimeHost(ctx, st, collector)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// Load Module
	svcID := "kv-service"
	version := "v1"
	if err := host.LoadModule(ctx, svcID, version, wasmBytes); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}
	if err := host.SetActiveVersion(ctx, svcID, version); err != nil {
		t.Fatalf("Failed to set active version: %v", err)
	}

	// Invoke
	// payload = "foo" (key to lookup)
	// We call "Handle" which is exported via //go:wasmexport Handle
	res, err := host.Invoke(ctx, svcID, "Handle", []byte("foo"))
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	expected := "value: bar"
	if string(res) != expected {
		t.Errorf("Expected %q, got %q", expected, string(res))
	}
}
