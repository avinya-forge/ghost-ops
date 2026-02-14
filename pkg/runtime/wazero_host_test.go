package runtime

import (
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
