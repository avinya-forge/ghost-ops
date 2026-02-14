package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"ghost-ops/pkg/protocol"
)

func TestJSONFileStore(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_store.json")

	store, err := NewJSONFileStore(filePath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	ctx := context.Background()

	// Initial check: empty
	services, err := store.ListServices(ctx)
	if err != nil {
		t.Fatalf("Failed to list services: %v", err)
	}
	if len(services) != 0 {
		t.Errorf("Expected 0 services, got %d", len(services))
	}

	// Add a service
	record := protocol.ServiceRecord{
		ServiceID:          "service-1",
		WASMHash:           "hash1",
		CurrentState:       protocol.StateActive,
		SynthesisTimestamp: time.Now(),
	}

	if err := store.UpdateService(ctx, record); err != nil {
		t.Fatalf("Failed to update service: %v", err)
	}

	// Verify it was added
	svc, err := store.GetService(ctx, "service-1")
	if err != nil {
		t.Fatalf("Failed to get service: %v", err)
	}
	if svc == nil {
		t.Fatal("Expected service to be found, got nil")
	}
	if svc.ServiceID != "service-1" {
		t.Errorf("Expected ServiceID 'service-1', got '%s'", svc.ServiceID)
	}

	// Update the service
	record.WASMHash = "hash2"
	if err := store.UpdateService(ctx, record); err != nil {
		t.Fatalf("Failed to update service: %v", err)
	}

	svc, err = store.GetService(ctx, "service-1")
	if err != nil {
		t.Fatalf("Failed to get service: %v", err)
	}
	if svc.WASMHash != "hash2" {
		t.Errorf("Expected WASMHash 'hash2', got '%s'", svc.WASMHash)
	}

	// List services again
	services, err = store.ListServices(ctx)
	if err != nil {
		t.Fatalf("Failed to list services: %v", err)
	}
	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}
}
