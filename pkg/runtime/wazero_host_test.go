package runtime

import (
	"bytes"
	"context"
	"testing"

	"ghost-ops/pkg/protocol"
)

func TestWazeroRuntimeHost(t *testing.T) {
	ctx := context.Background()
	store := protocol.NewInMemoryStateStore()
	host, err := NewWazeroRuntimeHost(ctx, store)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// Minimal WASM module exporting "test" function
	// (module (func (export "test")))
	wasm := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00,
		0x01, 0x04, 0x01, 0x60, 0x00, 0x00,
		0x03, 0x02, 0x01, 0x00,
		0x07, 0x08, 0x01, 0x04, 0x74, 0x65, 0x73, 0x74, 0x00, 0x00,
		0x0a, 0x04, 0x01, 0x02, 0x00, 0x0b,
	}

	serviceID := "test-service"

	// Test LoadModule
	if err := host.LoadModule(ctx, serviceID, wasm); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Test Invoke
	if _, err := host.Invoke(ctx, serviceID, "test", nil); err != nil {
		t.Fatalf("Failed to invoke method: %v", err)
	}

	// Test Invoke non-existent method
	if _, err := host.Invoke(ctx, serviceID, "missing", nil); err == nil {
		t.Error("Expected error for missing method, got nil")
	}

	// Test UnloadModule
	if err := host.UnloadModule(ctx, serviceID); err != nil {
		t.Fatalf("Failed to unload module: %v", err)
	}

	// Test Invoke after unload
	if _, err := host.Invoke(ctx, serviceID, "test", nil); err == nil {
		t.Error("Expected error after unload, got nil")
	}
}

func TestWazeroRuntimeHost_Payload(t *testing.T) {
	ctx := context.Background()
	store := protocol.NewInMemoryStateStore()
	host, err := NewWazeroRuntimeHost(ctx, store)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// WASM module with payload support (malloc, echo, free)
	// (module
	//   (memory (export "memory") 1)
	//   (func (export "malloc") (param i32) (result i32) i32.const 1024)
	//   (func (export "echo") (param i32 i32) (result i64) ... packed ptr/len ...)
	//   (func (export "free") (param i32) ...)
	// )
	wasmPayload := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00,
		0x01, 0x10, 0x03, 0x60, 0x01, 0x7f, 0x01, 0x7f, 0x60, 0x02, 0x7f, 0x7f, 0x01, 0x7e, 0x60, 0x01, 0x7f, 0x00,
		0x03, 0x04, 0x03, 0x00, 0x01, 0x02,
		0x05, 0x03, 0x01, 0x00, 0x01,
		0x07, 0x21, 0x04, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x02, 0x00, 0x06, 0x6d, 0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x00, 0x00, 0x04, 0x65, 0x63, 0x68, 0x6f, 0x00, 0x01, 0x04, 0x66, 0x72, 0x65, 0x65, 0x00, 0x02,
		0x0a, 0x17, 0x03, 0x05, 0x00, 0x41, 0x80, 0x08, 0x0b, 0x0c, 0x00, 0x20, 0x00, 0xad, 0x20, 0x01, 0xad, 0x42, 0x20, 0x86, 0x84, 0x0b, 0x02, 0x00, 0x0b,
	}

	serviceID := "echo-service"

	if err := host.LoadModule(ctx, serviceID, wasmPayload); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	input := []byte("hello world")
	output, err := host.Invoke(ctx, serviceID, "echo", input)
	if err != nil {
		t.Fatalf("Failed to invoke echo: %v", err)
	}

	if !bytes.Equal(output, input) {
		t.Errorf("Expected output %q, got %q", input, output)
	}
}
