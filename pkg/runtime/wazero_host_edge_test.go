package runtime

import "ghost-ops/pkg/config"
import (
	"context"
	"strings"
	"testing"

	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/telemetry"
)

func TestWazeroRuntimeHost_EdgeCases(t *testing.T) {
	ctx := context.Background()
	store := protocol.NewInMemoryStateStore()
	collector := telemetry.NewInMemoryCollector()
	host, err := NewWazeroRuntimeHost(ctx, store, collector, config.CapabilitiesConfig{})
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	t.Run("LoadModule_InvalidWASM", func(t *testing.T) {
		serviceID := "invalid-service"
		version := "v1"
		invalidWASM := []byte("this is not wasm")

		err := host.LoadModule(ctx, serviceID, version, invalidWASM)
		if err == nil {
			t.Fatal("Expected error for invalid WASM, got nil")
		}

		uniqueName := serviceID + "-" + version
		host.mu.RLock()
		_, exists := host.requests[uniqueName]
		host.mu.RUnlock()

		if exists {
			t.Errorf("Expected requests entry for %s to be cleaned up on failure, but it exists", uniqueName)
		}
	})

	t.Run("LoadModule_Duplicate", func(t *testing.T) {
		serviceID := "duplicate-service"
		version := "v1"
		// Minimal valid WASM module: magic header + version
		minimalWASM := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

		if err := host.LoadModule(ctx, serviceID, version, minimalWASM); err != nil {
			t.Fatalf("Failed to load valid minimal WASM: %v", err)
		}

		// Try loading again
		err := host.LoadModule(ctx, serviceID, version, minimalWASM)
		if err == nil {
			t.Fatal("Expected error for duplicate module load, got nil")
		}

		if !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "already loaded") {
			t.Errorf("Expected 'already exists' error, got: %v", err)
		}
	})

	t.Run("SetActiveVersion_NotLoaded", func(t *testing.T) {
		serviceID := "unknown-service"
		version := "v999"

		err := host.SetActiveVersion(ctx, serviceID, version)
		if err == nil {
			t.Fatal("Expected error for unknown version, got nil")
		}
	})

	t.Run("Invoke_NoActiveVersion", func(t *testing.T) {
		serviceID := "inactive-service"

		// Ensure no active version is set (it shouldn't be for a new ID)
		_, err := host.Invoke(ctx, serviceID, "method", []byte("payload"))
		if err == nil {
			t.Fatal("Expected error for invoke with no active version, got nil")
		}
		if !strings.Contains(err.Error(), "no active version") {
			t.Errorf("Expected 'no active version' error, got: %v", err)
		}
	})

	t.Run("Invoke_UnknownService", func(t *testing.T) {
		serviceID := "completely-unknown-service"

		_, err := host.Invoke(ctx, serviceID, "method", []byte("payload"))
		if err == nil {
			t.Fatal("Expected error for invoke on unknown service, got nil")
		}
	})
}
