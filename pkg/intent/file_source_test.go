package intent

import (
	"context"
	"encoding/json"
	"os"
	"testing"

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
