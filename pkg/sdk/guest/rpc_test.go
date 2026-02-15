package guest_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/runtime"
)

func TestGuestRPC(t *testing.T) {
	ctx := context.Background()
	engine := evolution.NewGoCompilerEngine()

	// Locate repo root (assuming we are in pkg/sdk/guest)
	// We need to write files inside the module for imports to work.
	// We'll use examples/services/rpc-test-generated
	cwd, _ := os.Getwd()
	repoRoot := filepath.Join(cwd, "../../..") // from pkg/sdk/guest
	testDir := filepath.Join(repoRoot, "examples", "services", "rpc-test-generated")

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Callee
	calleeDir := filepath.Join(testDir, "callee")
	if err := os.MkdirAll(calleeDir, 0755); err != nil {
		t.Fatalf("Failed to create callee dir: %v", err)
	}
	calleeSource := `package main
import "ghost-ops/pkg/sdk/guest"
func Handle(payload []byte) ([]byte, error) {
	return append([]byte("Echo: "), payload...), nil
}
func main() {
	guest.Register("Handle", Handle)
	guest.Start()
}
`
	calleeFile := filepath.Join(calleeDir, "main.go")
	if err := os.WriteFile(calleeFile, []byte(calleeSource), 0644); err != nil {
		t.Fatalf("Failed to write callee: %v", err)
	}

	// Caller
	callerDir := filepath.Join(testDir, "caller")
	if err := os.MkdirAll(callerDir, 0755); err != nil {
		t.Fatalf("Failed to create caller dir: %v", err)
	}
	callerSource := `package main
import "ghost-ops/pkg/sdk/guest"
func Call(payload []byte) ([]byte, error) {
	return guest.Call("Callee", "Handle", payload)
}
func main() {
	guest.Register("Call", Call)
	guest.Start()
}
`
	callerFile := filepath.Join(callerDir, "main.go")
	if err := os.WriteFile(callerFile, []byte(callerSource), 0644); err != nil {
		t.Fatalf("Failed to write caller: %v", err)
	}

	// Compile using source_path constraint
	calleeWasm, err := engine.Evolve(ctx, protocol.Blueprint{
		Constraints: map[string]interface{}{"source_path": calleeFile},
	})
	if err != nil {
		t.Fatalf("Failed to compile Callee: %v", err)
	}

	callerWasm, err := engine.Evolve(ctx, protocol.Blueprint{
		Constraints: map[string]interface{}{"source_path": callerFile},
	})
	if err != nil {
		t.Fatalf("Failed to compile Caller: %v", err)
	}

	// Setup Host
	// Use InMemoryStateStore from protocol package
	st := protocol.NewInMemoryStateStore()
	host, err := runtime.NewWazeroRuntimeHost(ctx, st)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// Load Services
	if err := host.LoadModule(ctx, "Callee", calleeWasm); err != nil {
		t.Fatalf("Failed to load Callee: %v", err)
	}
	if err := host.LoadModule(ctx, "Caller", callerWasm); err != nil {
		t.Fatalf("Failed to load Caller: %v", err)
	}

	// Invoke Caller
	payload := []byte("World")
	res, err := host.Invoke(ctx, "Caller", "Call", payload)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	expected := "Echo: World"
	if string(res) != expected {
		t.Errorf("Expected %q, got %q", expected, string(res))
	}
}
