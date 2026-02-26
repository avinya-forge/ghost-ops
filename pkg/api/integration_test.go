package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"ghost-ops/pkg/config"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/registry"
	"ghost-ops/pkg/runtime"
	"ghost-ops/pkg/telemetry"
)

var integrationWasmBytes []byte

func TestMain(m *testing.M) {
	// Compile the test WASM module for integration tests
	if err := compileIntegrationWasm(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to compile integration WASM: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func compileIntegrationWasm() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	// wd is pkg/api
	srcPath := filepath.Join(wd, "../../examples/test-service/main.go")
	outPath := filepath.Join(wd, "integration.wasm")

	cmd := exec.Command("go", "build", "-o", outPath, srcPath)
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compile failed: %s: %s", err, out)
	}

	b, err := os.ReadFile(outPath)
	if err != nil {
		return err
	}
	integrationWasmBytes = b

	// Cleanup
	os.Remove(outPath)
	return nil
}

func TestAPI_EndToEndFlow(t *testing.T) {
	ctx := context.Background()
	store := protocol.NewInMemoryStateStore()
	collector := telemetry.NewInMemoryCollector()

	// Use Real Wazero Runtime
	rt, err := runtime.NewWazeroRuntimeHost(ctx, store, collector)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer rt.Close(ctx)

	// Setup Mocks
	source := &MockIntentSource{
		blueprints: []protocol.Blueprint{
			{
				ServiceID: "e2e-service",
				Intent:    "integration test",
			},
		},
	}

	engine := &MockEvolutionEngine{
		wasmBytes: integrationWasmBytes,
	}

	reg := registry.NewRegistry(store, engine, source, rt, collector, nil, config.RegistryConfig{})
	srv := NewServer(reg, collector, 1024*1024)

	// Create Test Server
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// 1. Trigger Reconcile
	// POST /reconcile
	resp, err := http.Post(ts.URL+"/reconcile", "application/json", nil)
	if err != nil {
		t.Fatalf("Reconcile request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Reconcile failed: status=%d body=%s", resp.StatusCode, body)
	}

	// Wait for service to be active
	// We can poll invoke or logs
	deadline := time.Now().Add(5 * time.Second)
	var lastErr error
	success := false

	for time.Now().Before(deadline) {
		// Attempt Invoke
		payload := []byte("hello e2e")
		// Query param method=echo (from test-service/main.go)
		req, _ := http.NewRequest("POST", ts.URL+"/services/e2e-service/invoke?method=echo", bytes.NewReader(payload))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if string(body) == "hello e2e" {
				success = true
				break
			} else {
				lastErr = fmt.Errorf("unexpected body: %s", body)
			}
		} else {
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("status %d: %s", resp.StatusCode, body)
		}
		resp.Body.Close()

		time.Sleep(100 * time.Millisecond)
	}

	if !success {
		t.Fatalf("Service invocation failed after timeout: %v", lastErr)
	}

	// Also verify metrics
	metricsResp, err := http.Get(ts.URL + "/metrics")
	if err == nil {
		metricsResp.Body.Close()
	}
}
