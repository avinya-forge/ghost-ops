package guest_test

import "ghost-ops/pkg/config"
import (
	"context"
	"os"
	"testing"

	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/runtime"
	"ghost-ops/pkg/store"
	"ghost-ops/pkg/telemetry"
)

// Callee Source Code
const calleeSource = `
package main

import (
	"fmt"
	"ghost-ops/pkg/sdk/guest"
)

func Hello(payload []byte) ([]byte, error) {
	name := string(payload)
	return []byte(fmt.Sprintf("Hello, %s!", name)), nil
}

func main() {
	guest.Register("Hello", Hello)
	guest.Start()
}
`

// Caller Source Code
const callerSource = `
package main

import (
	"fmt"
	"ghost-ops/pkg/sdk/guest"
)

func InvokeCallee(payload []byte) ([]byte, error) {
	// payload is the name to send to callee
	res, err := guest.Call("callee", "Hello", payload)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("Callee said: %s", string(res))), nil
}

func main() {
	guest.Register("InvokeCallee", InvokeCallee)
	guest.Start()
}
`

func TestGuestSDK_RPC(t *testing.T) {
	ctx := context.Background()
	engine := evolution.NewGoCompilerEngine()

	// 1. Compile Callee
	calleeBP := protocol.Blueprint{
		Constraints: map[string]interface{}{
			"source_code": calleeSource,
		},
	}
	calleeWasm, err := engine.Evolve(ctx, calleeBP)
	if err != nil {
		t.Fatalf("Failed to compile callee: %v", err)
	}

	// 2. Compile Caller
	callerBP := protocol.Blueprint{
		Constraints: map[string]interface{}{
			"source_code": callerSource,
		},
	}
	callerWasm, err := engine.Evolve(ctx, callerBP)
	if err != nil {
		t.Fatalf("Failed to compile caller: %v", err)
	}

	// 3. Setup Host
	storePath := "rpc_test_store.json"
	st, err := store.NewJSONFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer os.Remove(storePath)

	collector := telemetry.NewInMemoryCollector()
	host, err := runtime.NewWazeroRuntimeHost(ctx, st, collector, config.CapabilitiesConfig{})
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// 4. Load Modules
	// Load Callee
	if err := host.LoadModule(ctx, "callee", "v1", calleeWasm); err != nil {
		t.Fatalf("Failed to load callee: %v", err)
	}
	if err := host.SetActiveVersion(ctx, "callee", "v1"); err != nil {
		t.Fatalf("Failed to activate callee: %v", err)
	}

	// Load Caller
	if err := host.LoadModule(ctx, "caller", "v1", callerWasm); err != nil {
		t.Fatalf("Failed to load caller: %v", err)
	}
	if err := host.SetActiveVersion(ctx, "caller", "v1"); err != nil {
		t.Fatalf("Failed to activate caller: %v", err)
	}

	// 5. Invoke Caller
	// We call "InvokeCallee" on "caller" service.
	res, err := host.Invoke(ctx, "caller", "InvokeCallee", []byte("World"))
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	expected := "Callee said: Hello, World!"
	if string(res) != expected {
		t.Errorf("Expected %q, got %q", expected, string(res))
	}
}
