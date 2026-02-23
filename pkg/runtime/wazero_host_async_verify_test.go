package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/telemetry"
)

var testWasmBytes []byte

func TestMain(m *testing.M) {
	// Compile the test WASM module
	if err := compileTestWasm(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to compile test WASM: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func compileTestWasm() error {
	// Locate source file
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	// Assuming wd is pkg/runtime
	srcPath := filepath.Join(wd, "../../examples/test-service/main.go")
	outPath := filepath.Join(wd, "test_async.wasm")

	cmd := exec.Command("go", "build", "-o", outPath, srcPath)
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compile failed: %s: %s", err, out)
	}

	// Read bytes
	b, err := os.ReadFile(outPath)
	if err != nil {
		return err
	}
	testWasmBytes = b

	// Cleanup
	os.Remove(outPath)
	return nil
}

// MockStore for runtime tests
type MockStore struct{}

func (m *MockStore) Get(ctx context.Context, key string) ([]byte, error) { return nil, nil }
func (m *MockStore) Set(ctx context.Context, key string, val []byte) error { return nil }
func (m *MockStore) GetService(ctx context.Context, id string) (*protocol.ServiceRecord, error) {
	return nil, nil
}
func (m *MockStore) UpdateService(ctx context.Context, r protocol.ServiceRecord) error { return nil }
func (m *MockStore) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	return nil, nil
}

func TestWazeroRuntimeHost_LoadModule_Async(t *testing.T) {
	ctx := context.Background()
	collector := telemetry.NewInMemoryCollector()
	host, err := NewWazeroRuntimeHost(ctx, &MockStore{}, collector)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	start := time.Now()
	err = host.LoadModule(ctx, "svc-async", "v1", testWasmBytes)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("LoadModule failed: %v", err)
	}

	// Compilation is synchronous, Instantiation is async.
	// We expect LoadModule to return relatively quickly, but compilation of a Go binary takes time (parsing WASM).
	// However, Wazero CompileModule is fast enough.
	// The key is that InstantiateModule (which starts the guest loop) is async.
	// If it was sync, it would block forever because the guest loop runs indefinitely.
	// So if LoadModule returns at all, it proves async instantiation!
	t.Logf("LoadModule took %v", duration)

	// If LoadModule blocked on guest Start(), it would timeout or hang.
	// Since we are here, it returned.
}

func TestWazeroRuntimeHost_Invoke_WaitsForModule(t *testing.T) {
	ctx := context.Background()
	collector := telemetry.NewInMemoryCollector()
	host, err := NewWazeroRuntimeHost(ctx, &MockStore{}, collector)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// Load
	if err := host.LoadModule(ctx, "svc-invoke", "v1", testWasmBytes); err != nil {
		t.Fatalf("LoadModule failed: %v", err)
	}

	// Set Active
	if err := host.SetActiveVersion(ctx, "svc-invoke", "v1"); err != nil {
		t.Fatalf("SetActiveVersion failed: %v", err)
	}

	// Invoke immediately
	// This should block until the guest is ready and processes the request.
	payload := []byte("hello")
	res, err := host.Invoke(ctx, "svc-invoke", "echo", payload)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if string(res) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", string(res))
	}
}
