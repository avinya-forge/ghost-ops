package runtime

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"

	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/telemetry"
)

func TestTracePropagation(t *testing.T) {
	// Setup Tracer
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	tracer := tp.Tracer("test-tracer")

	// Setup Runtime
	collector := telemetry.NewInMemoryCollector()
	store := protocol.NewInMemoryStateStore()
	host, err := NewWazeroRuntimeHost(context.Background(), store, collector)
	assert.NoError(t, err)
	defer host.Close(context.Background())

	// Guest Code that calls get_trace_context
	guestCode := `
package main

import (
	"unsafe"
)

//go:wasmimport ghost_ops get_trace_context
func get_trace_context(ptr, cap uint32) uint32

//go:wasmimport ghost_ops next_command
func next_command(methPtr, methCap uint32) uint64

//go:wasmimport ghost_ops submit_result
func submit_result(ptr, len uint32)

func main() {
	methBuf := make([]byte, 256)
	for {
		next_command(ptr(methBuf), uint32(cap(methBuf)))

		// Get Trace Context
		buf := make([]byte, 256)
		n := get_trace_context(ptr(buf), uint32(cap(buf)))
		traceCtx := string(buf[:n])

		out := []byte(traceCtx)
		submit_result(ptr(out), uint32(len(out)))
	}
}

func ptr(b []byte) uint32 {
	if len(b) == 0 {
		return 0
	}
	return uint32(uintptr(unsafe.Pointer(&b[0])))
}
`
	// Compile Guest
	wasmBytes, err := evolution.CompileGo(context.Background(), guestCode)
	if err != nil {
		t.Fatalf("Failed to compile guest: %v", err)
	}

	// Load Module
	serviceID := "trace-service"
	version := "1"
	err = host.LoadModule(context.Background(), serviceID, version, wasmBytes)
	assert.NoError(t, err)

	// Set Active
	err = host.SetActiveVersion(context.Background(), serviceID, version)
	assert.NoError(t, err)

	// Wait for module to be ready (async load)
	// We can retry Invoke until success or use CheckHealth loop
	assert.Eventually(t, func() bool {
		return host.CheckHealth(context.Background(), serviceID) == nil
	}, 5*time.Second, 100*time.Millisecond)

	// Start Span
	ctx, span := tracer.Start(context.Background(), "host-span")
	defer span.End()

	// Invoke
	payload := []byte("test")
	res, err := host.Invoke(ctx, serviceID, "test-method", payload)
	assert.NoError(t, err)

	// Verify Result matches "trace_id=...;span_id=..."
	resStr := string(res)
	expectedTraceID := span.SpanContext().TraceID().String()
	expectedSpanID := span.SpanContext().SpanID().String()

	assert.Contains(t, resStr, fmt.Sprintf("trace_id=%s", expectedTraceID))
	assert.Contains(t, resStr, fmt.Sprintf("span_id=%s", expectedSpanID))
}
