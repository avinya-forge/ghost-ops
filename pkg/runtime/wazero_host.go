package runtime

import (
	"context"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WazeroRuntimeHost implements RuntimeHost using wazero.
type WazeroRuntimeHost struct {
	runtime wazero.Runtime
	modules map[string]api.Module
	mu      sync.RWMutex
}

// NewWazeroRuntimeHost creates a new WazeroRuntimeHost.
func NewWazeroRuntimeHost(ctx context.Context) (*WazeroRuntimeHost, error) {
	r := wazero.NewRuntime(ctx)

	// Instantiate WASI, as many modules might need it.
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	return &WazeroRuntimeHost{
		runtime: r,
		modules: make(map[string]api.Module),
	}, nil
}

// LoadModule loads a WASM binary with a unique service ID.
func (h *WazeroRuntimeHost) LoadModule(ctx context.Context, serviceID string, wasmBytes []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.modules[serviceID]; exists {
		return fmt.Errorf("module %s already loaded", serviceID)
	}

	// Compile the module first.
	compiled, err := h.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("failed to compile module: %w", err)
	}
	// Note: We might want to close 'compiled' if we don't plan to re-instantiate it,
	// but wazero manages cache. Explicit close is good practice if we want to release memory,
	// but keeping it cached is also fine. For now, let's keep it simple.

	// Instantiate with the service ID as the module name.
	config := wazero.NewModuleConfig().WithName(serviceID)
	mod, err := h.runtime.InstantiateModule(ctx, compiled, config)
	if err != nil {
		return fmt.Errorf("failed to instantiate module: %w", err)
	}

	h.modules[serviceID] = mod
	return nil
}

// Invoke calls a function on the loaded module.
func (h *WazeroRuntimeHost) Invoke(ctx context.Context, serviceID, method string, payload []byte) ([]byte, error) {
	h.mu.RLock()
	mod, exists := h.modules[serviceID]
	h.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("module %s not found", serviceID)
	}

	fn := mod.ExportedFunction(method)
	if fn == nil {
		return nil, fmt.Errorf("function %s not exported", method)
	}

	// Call the function.
	results, err := fn.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("function call failed: %w", err)
	}

	// For now, return nil (or empty slice) as we don't have an ABI to decode results.
	_ = results
	return nil, nil
}

// UnloadModule removes a module.
func (h *WazeroRuntimeHost) UnloadModule(ctx context.Context, serviceID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	mod, exists := h.modules[serviceID]
	if !exists {
		return fmt.Errorf("module %s not found", serviceID)
	}

	if err := mod.Close(ctx); err != nil {
		return fmt.Errorf("failed to close module: %w", err)
	}

	delete(h.modules, serviceID)
	return nil
}

// Close closes the runtime.
func (h *WazeroRuntimeHost) Close(ctx context.Context) error {
	return h.runtime.Close(ctx)
}
