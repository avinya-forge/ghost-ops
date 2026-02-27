package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"

	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/telemetry"
)

func TestWazeroRuntimeHost_TraceContext(t *testing.T) {
	ctx := context.Background()
	store := protocol.NewInMemoryStateStore()
	collector := telemetry.NewInMemoryCollector()
	host, err := NewWazeroRuntimeHost(ctx, store, collector)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// Setup Tracer Provider
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(ctx)

	tracer := tp.Tracer("test-tracer")

	// Use the Go Compiler to build the kv-service example for testing
	// We will modify the source path to point to a service that uses GetTraceContext if we had one.
	// Since we don't have a dedicated test service for this yet, we will rely on creating a simple one or
	// assume that the infrastructure works if we can invoke a service.
	//
	// However, to properly test this, we need a WASM module that calls `guest.GetTraceContext`.
	// Let's create a temporary test service source file.

	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "main.go")
	srcContent := `
package main

import (
	"fmt"
	"ghost-ops/pkg/sdk/guest"
)

func main() {
	guest.Register("CheckTrace", func(payload []byte) ([]byte, error) {
		ctx, err := guest.GetTraceContext()
		if err != nil {
			return nil, err
		}
		traceID := ctx["trace_id"]
		spanID := ctx["span_id"]
		return []byte(fmt.Sprintf("%s:%s", traceID, spanID)), nil
	})
	guest.Start()
}
`
	if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
		t.Fatalf("Failed to write test source: %v", err)
	}

	engine := evolution.NewGoCompilerEngine()
	bp := protocol.Blueprint{
		Constraints: map[string]interface{}{
			"source_path": srcPath,
		},
	}

	wasmBytes, err := engine.Evolve(ctx, bp)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	serviceID := "trace-test-service"
	version := "v1"

	if err := host.LoadModule(ctx, serviceID, version, wasmBytes); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}
	if err := host.SetActiveVersion(ctx, serviceID, version); err != nil {
		t.Fatalf("Failed to set active version: %v", err)
	}

	// Start a Span
	ctx, span := tracer.Start(ctx, "test-span")
	defer span.End()

	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Invoke
	output, err := host.Invoke(ctx, serviceID, "CheckTrace", nil)
	if err != nil {
		t.Fatalf("Failed to invoke method: %v", err)
	}

	expected := fmt.Sprintf("%s:%s", traceID, spanID)
	if string(output) != expected {
		t.Errorf("Expected trace context %q, got %q", expected, string(output))
	}
}

// TestWazeroRuntimeHost_Propagation verifies that trace context is propagated across RPC calls.
func TestWazeroRuntimeHost_Propagation(t *testing.T) {
	ctx := context.Background()
	store := protocol.NewInMemoryStateStore()
	collector := telemetry.NewInMemoryCollector()
	host, err := NewWazeroRuntimeHost(ctx, store, collector)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// Setup Tracer
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(ctx)
	tracer := tp.Tracer("propagation-test")

	tempDir := t.TempDir()

	// Service A: Calls Service B
	srcA := filepath.Join(tempDir, "service_a.go")
	os.WriteFile(srcA, []byte(`
package main
import (
	"ghost-ops/pkg/sdk/guest"
)
func main() {
	guest.Register("CallB", func(p []byte) ([]byte, error) {
		return guest.Call("service-b", "GetTrace", nil)
	})
	guest.Start()
}
`), 0644)

	// Service B: Returns Trace Context
	srcB := filepath.Join(tempDir, "service_b.go")
	os.WriteFile(srcB, []byte(`
package main
import (
	"ghost-ops/pkg/sdk/guest"
)
func main() {
	guest.Register("GetTrace", func(p []byte) ([]byte, error) {
		ctx, _ := guest.GetTraceContext()
		return []byte(ctx["trace_id"]), nil
	})
	guest.Start()
}
`), 0644)

	engine := evolution.NewGoCompilerEngine()

	// Compile and Load A
	bpA := protocol.Blueprint{Constraints: map[string]interface{}{"source_path": srcA}}
	wasmA, errA := engine.Evolve(ctx, bpA)
	if errA != nil {
		t.Fatalf("Failed to evolve service A: %v", errA)
	}

	if err := host.LoadModule(ctx, "service-a", "v1", wasmA); err != nil {
		t.Fatalf("Failed to load service A: %v", err)
	}
	if err := host.SetActiveVersion(ctx, "service-a", "v1"); err != nil {
		t.Fatalf("Failed to activate service A: %v", err)
	}

	// Compile and Load B
	bpB := protocol.Blueprint{Constraints: map[string]interface{}{"source_path": srcB}}
	wasmB, errB := engine.Evolve(ctx, bpB)
	if errB != nil {
		t.Fatalf("Failed to evolve service B: %v", errB)
	}
	if err := host.LoadModule(ctx, "service-b", "v1", wasmB); err != nil {
		t.Fatalf("Failed to load service B: %v", err)
	}
	if err := host.SetActiveVersion(ctx, "service-b", "v1"); err != nil {
		t.Fatalf("Failed to activate service B: %v", err)
	}

	// Start Span
	ctx, span := tracer.Start(ctx, "root-span")
	defer span.End()
	traceID := span.SpanContext().TraceID().String()

	// Ensure propagation works (might need a small delay or retry if async instantiation is slow,
	// but host.LoadModule should block until loaded or at least until instantiation starts.
	// However, instantiation is async. We might need to wait for readiness.
	// But `LoadModule` in `wazero_host.go` launches a goroutine.
	// We need to wait for services to be ready.
	time.Sleep(100 * time.Millisecond)

	// Invoke A -> Call B
	output, err := host.Invoke(ctx, "service-a", "CallB", nil)
	if err != nil {
		t.Fatalf("Failed to invoke: %v", err)
	}

	if string(output) != traceID {
		t.Errorf("Trace ID mismatch. Expected %s, got %s", traceID, string(output))
	}
}
