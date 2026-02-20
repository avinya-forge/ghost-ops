package intent

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"ghost-ops/pkg/protocol"
)

func TestFileIntentSource(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "blueprints-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	blueprints := []protocol.Blueprint{
		{ServiceID: "svc-1", Intent: "intent-1"},
		{ServiceID: "svc-2", Intent: "intent-2"},
	}

	data, _ := json.Marshal(blueprints)
	if _, err := tmpfile.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test NewFileIntentSource
	source, err := NewFileIntentSource(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	// Test GetNextBlueprint
	ctx := context.Background()

	// 1st
	bp, err := source.GetNextBlueprint(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if bp == nil || bp.ServiceID != "svc-1" {
		t.Errorf("Expected svc-1, got %v", bp)
	}

	// 2nd
	bp, err = source.GetNextBlueprint(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if bp == nil || bp.ServiceID != "svc-2" {
		t.Errorf("Expected svc-2, got %v", bp)
	}

	// End
	bp, err = source.GetNextBlueprint(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if bp != nil {
		t.Errorf("Expected nil (end of list), got %v", bp)
	}
}

func TestFileIntentSource_Reload(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "blueprints-reload-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	blueprints1 := []protocol.Blueprint{
		{ServiceID: "svc-1", Intent: "intent-1"},
	}

	data1, _ := json.Marshal(blueprints1)
	if err := os.WriteFile(tmpfile.Name(), data1, 0644); err != nil {
		t.Fatal(err)
	}

	source, err := NewFileIntentSource(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	ctx := context.Background()

	// Read first blueprint
	bp, err := source.GetNextBlueprint(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if bp == nil || bp.ServiceID != "svc-1" {
		t.Errorf("Expected svc-1, got %v", bp)
	}

	// Read end (should be nil)
	bp, err = source.GetNextBlueprint(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if bp != nil {
		t.Errorf("Expected nil, got %v", bp)
	}

	// Update file with new blueprints
	// Ensure sufficient time passes for ModTime to change
	time.Sleep(1100 * time.Millisecond)

	blueprints2 := []protocol.Blueprint{
		{ServiceID: "svc-2", Intent: "intent-2"},
	}
	data2, _ := json.Marshal(blueprints2)
	if err := os.WriteFile(tmpfile.Name(), data2, 0644); err != nil {
		t.Fatal(err)
	}

	// Read next blueprint (should be svc-2 after reload)
	bp, err = source.GetNextBlueprint(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if bp == nil || bp.ServiceID != "svc-2" {
		t.Errorf("Expected svc-2 after reload, got %v", bp)
	}

	// Read end again
	bp, err = source.GetNextBlueprint(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if bp != nil {
		t.Errorf("Expected nil, got %v", bp)
	}
}

func TestFileIntentSource_EdgeCases(t *testing.T) {
	// Case 1: File does not exist
	_, err := NewFileIntentSource("non-existent-file.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	// Case 2: Empty File
	tmpfile, err := os.CreateTemp("", "blueprints-empty-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Create empty file source
	source, err := NewFileIntentSource(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to create source with empty file: %v", err)
	}

	ctx := context.Background()
	bp, err := source.GetNextBlueprint(ctx)
	if err != nil {
		t.Errorf("Unexpected error for empty file: %v", err)
	}
	if bp != nil {
		t.Errorf("Expected nil blueprint for empty file, got %v", bp)
	}

	// Case 3: Invalid JSON
	// Ensure we sleep enough to trigger reload
	time.Sleep(2 * time.Second)

	if err := os.WriteFile(tmpfile.Name(), []byte("{invalid-json"), 0644); err != nil {
		t.Fatal(err)
	}

	// We need to trigger reload by exhausting the list (it was already empty/exhausted)
	// GetNextBlueprint calls load() if index >= len
	_, err = source.GetNextBlueprint(ctx)
	if err == nil {
		t.Error("Expected error for invalid JSON reload, got nil")
	}

	// Case 4: Invalid JSON on initial load
	tmpfileInvalid, err := os.CreateTemp("", "blueprints-invalid-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfileInvalid.Name())
	if err := os.WriteFile(tmpfileInvalid.Name(), []byte("not-json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = NewFileIntentSource(tmpfileInvalid.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON on init, got nil")
	}
}
