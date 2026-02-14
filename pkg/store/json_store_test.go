package store

import (
	"context"
	"os"
	"testing"
	"time"

	"ghost-ops/pkg/protocol"
)

func TestJSONFileStore(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ghost-ops-store-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	store, err := NewJSONFileStore(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	ctx := context.Background()

	// Test UpdateService
	record := protocol.ServiceRecord{
		ServiceID:          "test-service-1",
		WASMHash:           "abc123hash",
		CurrentState:       protocol.StateActive,
		SynthesisTimestamp: time.Now().UTC(),
	}

	if err := store.UpdateService(ctx, record); err != nil {
		t.Fatalf("UpdateService failed: %v", err)
	}

	// Test GetService
	got, err := store.GetService(ctx, "test-service-1")
	if err != nil {
		t.Fatalf("GetService failed: %v", err)
	}
	if got == nil {
		t.Fatal("GetService returned nil")
	}
	if got.ServiceID != record.ServiceID {
		t.Errorf("Expected ServiceID %s, got %s", record.ServiceID, got.ServiceID)
	}
	if got.WASMHash != record.WASMHash {
		t.Errorf("Expected WASMHash %s, got %s", record.WASMHash, got.WASMHash)
	}

	// Test ListServices
	list, err := store.ListServices(ctx)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("Expected 1 service, got %d", len(list))
	}

	// Test Set
	if err := store.Set(ctx, "config-key", []byte("config-value")); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Test Get
	val, err := store.Get(ctx, "config-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(val) != "config-value" {
		t.Errorf("Expected 'config-value', got '%s'", string(val))
	}

	// Test persistence (reload store)
	store2, err := NewJSONFileStore(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create store2: %v", err)
	}
	val2, err := store2.Get(ctx, "config-key")
	if err != nil {
		t.Fatalf("Get from new store instance failed: %v", err)
	}
	if string(val2) != "config-value" {
		t.Errorf("Expected 'config-value' from new store, got '%s'", string(val2))
	}
}
