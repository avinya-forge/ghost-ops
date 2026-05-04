package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

	// Test Set
	if err := store.Set(ctx, "config-key", []byte("config-value")); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Test Get
	val, err := store.Get(ctx, "config-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(val) != "config-value" {
		t.Errorf("Expected 'config-value', got '%s'", string(val))
	}

	// Test persistence (reload store)
	store2, err := NewJSONFileStore(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create store2: %v", err)
	}
	val2, err := store2.Get(ctx, "config-key")
	if err != nil {
		t.Fatalf("Get from new store instance failed: %v", err)
	}
	if string(val2) != "config-value" {
		t.Errorf("Expected 'config-value' from new store, got '%s'", string(val2))
	}
}

// TestJSONFileStore_SaveIsAtomic ensures save() does not leave a half-written
// file visible at s.path. Regression test for BUG-053.
func TestJSONFileStore_SaveIsAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	store, err := NewJSONFileStore(path)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	ctx := context.Background()

	// Hammer with concurrent writes; at every read, the on-disk file must be
	// either empty initial-state JSON or fully-formed JSON containing N services.
	const writers = 8
	const writes = 50
	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < writes; i++ {
				rec := protocol.ServiceRecord{
					ServiceID:    "svc-" + string(rune('a'+w)),
					WASMHash:     strings.Repeat("a", 32),
					CurrentState: protocol.StateActive,
				}
				if err := store.UpdateService(ctx, rec); err != nil {
					t.Errorf("write w=%d i=%d: %v", w, i, err)
				}
			}
		}(w)
	}

	// Concurrent readers: at every observed moment the file on disk must parse.
	stop := make(chan struct{})
	var readerErr error
	var readerMu sync.Mutex
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			data, err := os.ReadFile(path)
			if err != nil {
				continue // file briefly absent during rename is OK
			}
			if len(data) == 0 {
				readerMu.Lock()
				readerErr = err
				readerMu.Unlock()
				continue
			}
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				readerMu.Lock()
				readerErr = err
				readerMu.Unlock()
				return
			}
		}
	}()

	wg.Wait()
	close(stop)
	readerMu.Lock()
	defer readerMu.Unlock()
	if readerErr != nil {
		t.Fatalf("reader observed half-written file: %v", readerErr)
	}

	// Verify no orphaned .store-*.tmp files were left behind.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".store-") && strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("orphaned tmp file left behind: %s", e.Name())
		}
	}
}

// TestJSONFileStore_SaveLeavesGoodFileOnError verifies that a save failure
// does not corrupt the previously-good file. We simulate failure by pointing
// the store at a file inside a non-existent directory after the first write.
func TestJSONFileStore_SaveLeavesGoodFileOnError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	store, err := NewJSONFileStore(path)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	ctx := context.Background()

	good := protocol.ServiceRecord{ServiceID: "good", WASMHash: "h"}
	if err := store.UpdateService(ctx, good); err != nil {
		t.Fatalf("first write: %v", err)
	}
	contentBefore, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read before: %v", err)
	}

	// Force the next save to fail by pointing the store at a path whose
	// parent directory does not exist. CreateTemp will fail; the original
	// file (at the old path) remains untouched.
	store.path = filepath.Join(dir, "does-not-exist", "store.json")

	if err := store.UpdateService(ctx, protocol.ServiceRecord{ServiceID: "bad"}); err == nil {
		t.Fatal("expected UpdateService to fail when target directory is missing")
	}

	// File at the original good path must be byte-identical and still valid.
	contentAfter, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after: %v", err)
	}
	if string(contentBefore) != string(contentAfter) {
		t.Fatalf("good file changed despite save error\nbefore=%s\nafter =%s",
			contentBefore, contentAfter)
	}
	if !json.Valid(contentAfter) {
		t.Fatalf("file is no longer valid JSON: %s", contentAfter)
	}

	_ = time.Second
}
