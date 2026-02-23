package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"ghost-ops/pkg/telemetry"
	"github.com/tetratelabs/wazero/api"
)

// mockCompiledModule implements wazero.CompiledModule for testing.
type mockCompiledModule struct {
	name string
}

func (m *mockCompiledModule) Name() string { return m.name }
func (m *mockCompiledModule) ImportedFunctions() []api.FunctionDefinition { return nil }
func (m *mockCompiledModule) ExportedFunctions() map[string]api.FunctionDefinition { return nil }
func (m *mockCompiledModule) Close(context.Context) error { return nil }
func (m *mockCompiledModule) ImportedMemories() []api.MemoryDefinition { return nil }
func (m *mockCompiledModule) ExportedMemories() map[string]api.MemoryDefinition { return nil }
func (m *mockCompiledModule) CustomSections() []api.CustomSection { return nil }

func TestWazeroRuntimeHost_Cache_Hit(t *testing.T) {
	ctx := context.Background()
	collector := telemetry.NewInMemoryCollector()
	host, err := NewWazeroRuntimeHost(ctx, &MockStore{}, collector)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// 1. First Load (Miss)
	err = host.LoadModule(ctx, "svc-cache", "v1", testWasmBytes)
	if err != nil {
		t.Fatalf("LoadModule 1 failed: %v", err)
	}

	// Check Metric Miss
	snapshot := collector.Snapshot()
	// Depending on implementation, snapshot might return int64 or something else.
	// We cast to int64 or check generic equality
	// Also checking if key exists
	if val, ok := snapshot["module_cache_miss{service_id=svc-cache}"]; !ok {
		t.Errorf("Expected cache miss metric, got none")
	} else if val != int64(1) {
		t.Errorf("Expected cache miss 1, got %v", val)
	}

	// 2. Second Load (Hit) - Same WASM
	// We use different service ID to allow loading (since same uniqueName is blocked)
	err = host.LoadModule(ctx, "svc-cache-2", "v1", testWasmBytes)
	if err != nil {
		t.Fatalf("LoadModule 2 failed: %v", err)
	}

	snapshot = collector.Snapshot()
	if val, ok := snapshot["module_cache_hit{service_id=svc-cache-2}"]; !ok {
		t.Errorf("Expected cache hit metric, got none")
	} else if val != int64(1) {
		t.Errorf("Expected cache hit 1, got %v", val)
	}
}

func TestWazeroRuntimeHost_Cache_Eviction(t *testing.T) {
	ctx := context.Background()
	collector := telemetry.NewInMemoryCollector()
	host, err := NewWazeroRuntimeHost(ctx, &MockStore{}, collector)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close(ctx)

	// Manually fill cache with 50 items
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("hash-%d", i)
		host.compiledCache[key] = &mockCompiledModule{name: key}
		host.cacheOrder = append(host.cacheOrder, key)
	}

	if len(host.compiledCache) != 50 {
		t.Fatalf("Setup failed, count=%d", len(host.compiledCache))
	}

	// Oldest is "hash-0"
	if host.cacheOrder[0] != "hash-0" {
		t.Errorf("Expected oldest to be hash-0, got %s", host.cacheOrder[0])
	}

	// Load Real Module
	// This generates a hash different from "hash-X"
	err = host.LoadModule(ctx, "svc-evict", "v1", testWasmBytes)
	if err != nil {
		t.Fatalf("LoadModule failed: %v", err)
	}

	// Cache size should still be 50
	if len(host.compiledCache) != 50 {
		t.Errorf("Expected cache size 50, got %d", len(host.compiledCache))
	}

	// Check if "hash-0" is gone
	if _, ok := host.compiledCache["hash-0"]; ok {
		t.Errorf("Expected hash-0 to be evicted")
	}

	// Check if "hash-1" is new oldest
	if host.cacheOrder[0] != "hash-1" {
		t.Errorf("Expected oldest to be hash-1, got %s", host.cacheOrder[0])
	}

	// Check if real module is in cache (newest)
	hash := sha256.Sum256(testWasmBytes)
	realHash := hex.EncodeToString(hash[:])

	if host.cacheOrder[len(host.cacheOrder)-1] != realHash {
		t.Errorf("Expected newest to be %s, got %s", realHash, host.cacheOrder[len(host.cacheOrder)-1])
	}
}
