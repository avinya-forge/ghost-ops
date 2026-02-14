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
}
