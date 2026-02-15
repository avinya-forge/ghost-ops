package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"ghost-ops/pkg/protocol"
)

// Request represents a pending invocation.
type Request struct {
	method     string
	payload    []byte
	responseCh chan Response
}

// Response represents the result of an invocation.
type Response struct {
	payload []byte
	err     error
}

// WazeroRuntimeHost implements RuntimeHost using wazero.
type WazeroRuntimeHost struct {
	runtime    wazero.Runtime
	modules    map[string]api.Module
	requests   map[string]chan Request
	currentReq map[string]Request
	mu         sync.RWMutex
}

// NewWazeroRuntimeHost creates a new WazeroRuntimeHost.
func NewWazeroRuntimeHost(ctx context.Context, store protocol.StateStore) (*WazeroRuntimeHost, error) {
	r := wazero.NewRuntime(ctx)

	// Instantiate WASI, as many modules might need it.
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	h := &WazeroRuntimeHost{
		runtime:    r,
		modules:    make(map[string]api.Module),
		requests:   make(map[string]chan Request),
		currentReq: make(map[string]Request),
	}

	// Define kv_get host function
	kvGet := func(ctx context.Context, m api.Module, keyPtr, keyLen, valPtr, valLen uint32) uint32 {
		keyBytes, ok := m.Memory().Read(keyPtr, keyLen)
		if !ok {
			return 0
		}
		key := string(keyBytes)

		val, err := store.Get(ctx, key)
		if err != nil {
			return 0
		}

		if uint32(len(val)) > valLen {
			if !m.Memory().Write(valPtr, val[:valLen]) {
				return 0
			}
		} else {
			if !m.Memory().Write(valPtr, val) {
				return 0
			}
		}

		return uint32(len(val))
	}

	// Define rpc host function
	rpc := func(ctx context.Context, m api.Module, svcPtr, svcLen, methPtr, methLen, payPtr, payLen, outPtr, outLen uint32) uint32 {
		mem := m.Memory()
		svcBytes, ok := mem.Read(svcPtr, svcLen)
		if !ok { return 0 }
		serviceID := string(svcBytes)

		methBytes, ok := mem.Read(methPtr, methLen)
		if !ok { return 0 }
		method := string(methBytes)

		payload, ok := mem.Read(payPtr, payLen)
		if !ok { return 0 }
		payloadCopy := make([]byte, len(payload))
		copy(payloadCopy, payload)

		result, err := h.Invoke(ctx, serviceID, method, payloadCopy)
		if err != nil { return 0 }

		if uint32(len(result)) > outLen {
			if !mem.Write(outPtr, result[:outLen]) { return 0 }
		} else {
			if !mem.Write(outPtr, result) { return 0 }
		}

		return uint32(len(result))
	}

	// Define next_command host function
	nextCommand := func(ctx context.Context, m api.Module, methPtr, methCap uint32) uint64 {
		serviceID := m.Name()

		h.mu.Lock()
		if _, known := h.modules[serviceID]; !known {
			h.modules[serviceID] = m
		}
		reqCh, ok := h.requests[serviceID]
		h.mu.Unlock()

		if !ok {
			return 0
		}

		select {
		case req := <-reqCh:
			h.mu.Lock()
			h.currentReq[serviceID] = req
			h.mu.Unlock()

			methBytes := []byte(req.method)
			if uint32(len(methBytes)) > methCap {
				methBytes = methBytes[:methCap]
			}
			m.Memory().Write(methPtr, methBytes)

			return (uint64(len(req.payload)) << 32) | uint64(len(methBytes))

		case <-ctx.Done():
			return 0
		}
	}

	// Define read_payload host function
	readPayload := func(ctx context.Context, m api.Module, ptr, cap uint32) {
		serviceID := m.Name()
		h.mu.RLock()
		req, ok := h.currentReq[serviceID]
		h.mu.RUnlock()
		if !ok { return }

		if uint32(len(req.payload)) > cap {
			m.Memory().Write(ptr, req.payload[:cap])
		} else {
			m.Memory().Write(ptr, req.payload)
		}
	}

	// Define submit_result host function
	submitResult := func(ctx context.Context, m api.Module, ptr, len uint32) {
		serviceID := m.Name()
		h.mu.RLock()
		req, ok := h.currentReq[serviceID]
		h.mu.RUnlock()
		if !ok { return }

		var res []byte
		if len > 0 {
			resBytes, ok := m.Memory().Read(ptr, len)
			if ok {
				res = make([]byte, len)
				copy(res, resBytes)
			}
		}

		// Cleanup before sending response
		h.mu.Lock()
		delete(h.currentReq, serviceID)
		h.mu.Unlock()

		select {
		case req.responseCh <- Response{payload: res}:
		default:
		}
	}

	// Instantiate host module
	_, err := r.NewHostModuleBuilder("ghost_ops").
		NewFunctionBuilder().WithFunc(kvGet).Export("kv_get").
		NewFunctionBuilder().WithFunc(rpc).Export("rpc").
		NewFunctionBuilder().WithFunc(nextCommand).Export("next_command").
		NewFunctionBuilder().WithFunc(readPayload).Export("read_payload").
		NewFunctionBuilder().WithFunc(submitResult).Export("submit_result").
		Instantiate(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to instantiate host module: %w", err)
	}

	return h, nil
}

// LoadModule loads a WASM binary with a unique service ID.
func (h *WazeroRuntimeHost) LoadModule(ctx context.Context, serviceID string, wasmBytes []byte) error {
	h.mu.Lock()
	if _, exists := h.modules[serviceID]; exists {
		h.mu.Unlock()
		return fmt.Errorf("module %s already loaded", serviceID)
	}

	h.requests[serviceID] = make(chan Request)
	h.mu.Unlock()

	compiled, err := h.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("failed to compile module: %w", err)
	}

	// Instantiate in goroutine
	go func() {
		config := wazero.NewModuleConfig().WithName(serviceID)
		_, err := h.runtime.InstantiateModule(ctx, compiled, config)
		if err != nil {
			slog.Error("Failed to instantiate module", "service_id", serviceID, "error", err)
		}
	}()

	return nil
}

// Invoke calls a function on the loaded module.
func (h *WazeroRuntimeHost) Invoke(ctx context.Context, serviceID, method string, payload []byte) ([]byte, error) {
	h.mu.RLock()
	reqCh, exists := h.requests[serviceID]
	h.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("module %s not found or not ready", serviceID)
	}

	responseCh := make(chan Response, 1) // Buffered
	req := Request{
		method:     method,
		payload:    payload,
		responseCh: responseCh,
	}

	select {
	case reqCh <- req:
		// Request sent
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case res := <-responseCh:
		return res.payload, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// UnloadModule removes a module.
func (h *WazeroRuntimeHost) UnloadModule(ctx context.Context, serviceID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Ensure module is closed using runtime reference if possible
	if mod := h.runtime.Module(serviceID); mod != nil {
		if err := mod.Close(ctx); err != nil {
			slog.Warn("Failed to close module via runtime", "service_id", serviceID, "error", err)
		}
	} else if mod, exists := h.modules[serviceID]; exists {
		// Fallback
		mod.Close(ctx)
	}

	delete(h.modules, serviceID)
	delete(h.requests, serviceID)
	delete(h.currentReq, serviceID)

	return nil
}

// Close closes the runtime.
func (h *WazeroRuntimeHost) Close(ctx context.Context) error {
	return h.runtime.Close(ctx)
}
