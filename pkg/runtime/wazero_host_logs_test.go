package runtime

import (
	"context"
	"strings"
	"testing"
	"time"

	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/telemetry"
)

func TestWazeroRuntimeHost_GetLogs(t *testing.T) {
	ctx := context.Background()
	store := protocol.NewInMemoryStateStore()
	collector := telemetry.NewInMemoryCollector()
	host, err := NewWazeroRuntimeHost(ctx, store, collector)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// Simple Go program that prints to stdout/stderr
	source := `package main
import (
	"fmt"
	"os"
	"time"
)
func main() {
	fmt.Println("This is stdout")
	fmt.Fprintln(os.Stderr, "This is stderr")
	// Keep running briefly to simulate a service
	time.Sleep(100 * time.Millisecond)
}
`
	// Compile
	wasmBytes, err := evolution.CompileGo(ctx, source)
	if err != nil {
		t.Fatalf("Failed to compile source: %v", err)
	}

	serviceID := "log-test-service"
	version := "v1"

	// Load
	if err := host.LoadModule(ctx, serviceID, version, wasmBytes); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Set Active
	if err := host.SetActiveVersion(ctx, serviceID, version); err != nil {
		t.Fatalf("Failed to set active version: %v", err)
	}

	// Wait a bit for the goroutine to run main() and print
	time.Sleep(500 * time.Millisecond)

	// Get Logs
	logs, err := host.GetLogs(ctx, serviceID)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	logStr := string(logs)
	if !strings.Contains(logStr, "This is stdout") {
		t.Errorf("Logs missing stdout. Got: %s", logStr)
	}
	if !strings.Contains(logStr, "This is stderr") {
		t.Errorf("Logs missing stderr. Got: %s", logStr)
	}
}

func TestWazeroRuntimeHost_GetLogs_NoActiveVersion(t *testing.T) {
	ctx := context.Background()
	store := protocol.NewInMemoryStateStore()
	collector := telemetry.NewInMemoryCollector()
	host, err := NewWazeroRuntimeHost(ctx, store, collector)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	_, err = host.GetLogs(ctx, "non-existent")
	if err == nil {
		t.Fatal("Expected error for non-existent service, got nil")
	}
	if !strings.Contains(err.Error(), "no active version") {
		t.Errorf("Expected 'no active version' error, got: %v", err)
	}
}
