package runtime

import (
	"context"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"ghost-ops/pkg/protocol"
)

// WazeroRuntimeHost implements RuntimeHost using wazero.
type WazeroRuntimeHost struct {
	runtime wazero.Runtime
	modules map[string]api.Module
	mu      sync.RWMutex
}

// NewWazeroRuntimeHost creates a new WazeroRuntimeHost.
func NewWazeroRuntimeHost(ctx context.Context, store protocol.StateStore) (*WazeroRuntimeHost, error) {
	r := wazero.NewRuntime(ctx)

	// Instantiate WASI, as many modules might need it.
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	h := &WazeroRuntimeHost{
		runtime: r,
		modules: make(map[string]api.Module),
	}

	// Define kv_get host function
	// kv_get(keyPtr, keyLen, valPtr, valLen) -> valLen (or 0 if error/not found)
	kvGet := func(ctx context.Context, m api.Module, keyPtr, keyLen, valPtr, valLen uint32) uint32 {
		// Read key from memory
		keyBytes, ok := m.Memory().Read(keyPtr, keyLen)
		if !ok {
			return 0
		}
		key := string(keyBytes)

		// Get value from store
		val, err := store.Get(ctx, key)
		if err != nil {
			return 0 // Return 0 on error or not found
		}

		// Write value to memory
		toWrite := val
		if uint32(len(val)) > valLen {
			toWrite = val[:valLen]
		}
		if !m.Memory().Write(valPtr, toWrite) {
			return 0
		}

		return uint32(len(val))
	}

	// Define rpc host function
	// rpc(svcPtr, svcLen, methPtr, methLen, payPtr, payLen, outPtr, outLen) -> outLen
	rpc := func(ctx context.Context, m api.Module, svcPtr, svcLen, methPtr, methLen, payPtr, payLen, outPtr, outLen uint32) uint32 {
		mem := m.Memory()

		// Read Service ID
		svcBytes, ok := mem.Read(svcPtr, svcLen)
		if !ok {
			return 0
		}
		serviceID := string(svcBytes)

		// Read Method
		methBytes, ok := mem.Read(methPtr, methLen)
		if !ok {
			return 0
		}
		method := string(methBytes)

		// Read Payload
		payload, ok := mem.Read(payPtr, payLen)
		if !ok {
			return 0
		}
		// Make a copy of payload since we pass it to Invoke and memory might change?
		// Invoke takes []byte. It should be fine to pass the slice directly if Invoke uses it immediately.
		// However, to be safe, copy it.
		payloadCopy := make([]byte, len(payload))
		copy(payloadCopy, payload)

		// Invoke
		// Note: We use the context from the call, which should have the timeout/cancel from the original request.
		result, err := h.Invoke(ctx, serviceID, method, payloadCopy)
		if err != nil {
			return 0
		}

		// Write result
		toWrite := result
		if uint32(len(result)) > outLen {
			toWrite = result[:outLen]
		}
		if !mem.Write(outPtr, toWrite) {
			return 0
		}

		return uint32(len(result))
	}

	// Instantiate host module
	_, err := r.NewHostModuleBuilder("ghost_ops").
		NewFunctionBuilder().
		WithFunc(kvGet).
		Export("kv_get").
		NewFunctionBuilder().
		WithFunc(rpc).
		Export("rpc").
		Instantiate(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to instantiate host module: %w", err)
	}

	return h, nil
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

	paramTypes := fn.Definition().ParamTypes()

	var results []uint64
	var errCall error

	if len(paramTypes) == 0 {
		// Legacy support: 0 arguments
		results, errCall = fn.Call(ctx)
	} else if len(paramTypes) == 2 {
		// Payload passing support: (ptr, len)
		// We expect both parameters to be i32.
		if paramTypes[0] != api.ValueTypeI32 || paramTypes[1] != api.ValueTypeI32 {
			return nil, fmt.Errorf("function %s has 2 parameters but they are not (i32, i32)", method)
		}

		// We need to allocate memory in the module.
		malloc := mod.ExportedFunction("malloc")
		if malloc == nil {
			return nil, fmt.Errorf("function %s requires payload but module does not export malloc", method)
		}

		payloadLen := uint64(len(payload))
		// Call malloc(size) -> ptr
		res, err := malloc.Call(ctx, payloadLen)
		if err != nil {
			return nil, fmt.Errorf("malloc failed: %w", err)
		}
		if len(res) == 0 {
			return nil, fmt.Errorf("malloc returned no result")
		}
		ptr := res[0]

		// Write payload to memory
		if !mod.Memory().Write(uint32(ptr), payload) {
			return nil, fmt.Errorf("failed to write payload to memory at %d", ptr)
		}

		// Call function(ptr, len)
		results, errCall = fn.Call(ctx, ptr, payloadLen)

		// Free input memory if free is exported
		free := mod.ExportedFunction("free")
		if free != nil {
			if _, err := free.Call(ctx, ptr); err != nil {
				return nil, fmt.Errorf("failed to free input memory: %w", err)
			}
		}
	} else {
		return nil, fmt.Errorf("function %s has unsupported parameter count: %d (expected 0 or 2)", method, len(paramTypes))
	}

	if errCall != nil {
		return nil, fmt.Errorf("function call failed: %w", errCall)
	}

	// Handle return values.
	// Expecting 1 result: i64 (packed ptr/len)
	// high 32 = len, low 32 = ptr
	if len(results) == 1 {
		val := results[0]
		// Assuming i64 packed: ptr | (len << 32)
		// Note: This assumes Little Endian packing logic or standard bitwise ops.
		// If the WASM returns i64, it's just a value.
		resPtr := uint32(val)
		resLen := uint32(val >> 32)

		if resLen == 0 {
			return nil, nil
		}

		bytes, ok := mod.Memory().Read(resPtr, resLen)
		if !ok {
			return nil, fmt.Errorf("failed to read result from memory at %d len %d", resPtr, resLen)
		}

		// Return a copy
		out := make([]byte, len(bytes))
		copy(out, bytes)

		// Free result memory if free is exported
		free := mod.ExportedFunction("free")
		if free != nil && resPtr != 0 {
			if _, err := free.Call(ctx, uint64(resPtr)); err != nil {
				return nil, fmt.Errorf("failed to free result memory: %w", err)
			}
		}

		return out, nil
	}

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
