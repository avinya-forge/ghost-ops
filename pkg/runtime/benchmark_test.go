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

func BenchmarkInvoke(b *testing.B) {
	ctx := context.Background()
	stateStore := protocol.NewInMemoryStateStore()
	collector := telemetry.NewInMemoryCollector()
	host, err := NewWazeroRuntimeHost(ctx, stateStore, collector, config.CapabilitiesConfig{})
	if err != nil {
		b.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// Compile the kv-service example
	engine := evolution.NewGoCompilerEngine()
	// Adjust path relative to pkg/runtime
	srcPath, err := filepath.Abs("../../examples/services/kv-service/main.go")
	if err != nil {
		b.Fatalf("Failed to resolve source path: %v", err)
	}

	if _, err := os.Stat(srcPath); err != nil {
		b.Fatalf("Source file not found at %s: %v", srcPath, err)
	}

	bp := protocol.Blueprint{
		Constraints: map[string]interface{}{
			"source_path": srcPath,
		},
	}

	wasmBytes, err := engine.Evolve(ctx, bp)
	if err != nil {
		b.Fatalf("Failed to compile: %v", err)
	}

	serviceID := "bench-service"
	version := "v1"

	if err := host.LoadModule(ctx, serviceID, version, wasmBytes); err != nil {
		b.Fatalf("Failed to load module: %v", err)
	}

	if err := host.SetActiveVersion(ctx, serviceID, version); err != nil {
		b.Fatalf("Failed to activate version: %v", err)
	}

	// Set initial value
	if err := stateStore.Set(ctx, "bench_key", []byte("bench_value")); err != nil {
		b.Fatalf("Failed to set initial value: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// kv-service expects payload to be key name
		_, err := host.Invoke(ctx, serviceID, "Handle", []byte("bench_key"))
		if err != nil {
			b.Fatalf("Invoke failed: %v", err)
		}
	}
}
