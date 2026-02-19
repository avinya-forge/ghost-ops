package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"ghost-ops/pkg/protocol"
)

// request represents a pending invocation.
type request struct {
	method     string
	payload    []byte
	responseCh chan response
}

// response represents the result of an invocation.
type response struct {
	payload []byte
	err     error
}

// WazeroRuntimeHost implements RuntimeHost using wazero.
type WazeroRuntimeHost struct {
	runtime        wazero.Runtime
	modules        map[string]api.Module        // Key: uniqueName (serviceID-version)
	requests       map[string]chan request      // Key: uniqueName
	currentReq     map[string]request           // Key: uniqueName
	activeVersions map[string]string            // Key: serviceID, Value: uniqueName
	shadowVersions map[string]string            // Key: serviceID, Value: uniqueName
	mu             sync.RWMutex
	collector      protocol.MetricsCollector
}

// NewWazeroRuntimeHost creates a new WazeroRuntimeHost.
func NewWazeroRuntimeHost(ctx context.Context, store protocol.StateStore, collector protocol.MetricsCollector) (*WazeroRuntimeHost, error) {
	r := wazero.NewRuntime(ctx)

	// Instantiate WASI, as many modules might need it.
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	h := &WazeroRuntimeHost{
		runtime:        r,
		modules:        make(map[string]api.Module),
		requests:       make(map[string]chan request),
		currentReq:     make(map[string]request),
		activeVersions: make(map[string]string),
		shadowVersions: make(map[string]string),
		collector:      collector,
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
		targetServiceID := string(svcBytes)

		methBytes, ok := mem.Read(methPtr, methLen)
		if !ok { return 0 }
		method := string(methBytes)

		payload, ok := mem.Read(payPtr, payLen)
		if !ok { return 0 }
		payloadCopy := make([]byte, len(payload))
		copy(payloadCopy, payload)

		result, err := h.Invoke(ctx, targetServiceID, method, payloadCopy)
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
		moduleName := m.Name()

		h.mu.Lock()
		if _, known := h.modules[moduleName]; !known {
			h.modules[moduleName] = m
		}
		reqCh, ok := h.requests[moduleName]
		h.mu.Unlock()

		if !ok {
			return 0
		}

		select {
		case req := <-reqCh:
			h.mu.Lock()
			h.currentReq[moduleName] = req
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
		moduleName := m.Name()
		h.mu.RLock()
		req, ok := h.currentReq[moduleName]
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
		moduleName := m.Name()
		h.mu.RLock()
		req, ok := h.currentReq[moduleName]
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
		delete(h.currentReq, moduleName)
		h.mu.Unlock()

		select {
		case req.responseCh <- response{payload: res}:
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

// LoadModule loads a WASM binary for a specific version.
func (h *WazeroRuntimeHost) LoadModule(ctx context.Context, serviceID, version string, wasmBytes []byte) error {
	uniqueName := fmt.Sprintf("%s-%s", serviceID, version)

	h.mu.Lock()
	if _, exists := h.modules[uniqueName]; exists {
		h.mu.Unlock()
		return fmt.Errorf("module %s already loaded", uniqueName)
	}
	// Also check if channel exists
	if _, exists := h.requests[uniqueName]; exists {
		h.mu.Unlock()
		return fmt.Errorf("request channel for %s already exists", uniqueName)
	}

	h.requests[uniqueName] = make(chan request)
	h.mu.Unlock()

	compiled, err := h.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		h.mu.Lock()
		delete(h.requests, uniqueName)
		h.mu.Unlock()
		return fmt.Errorf("failed to compile module: %w", err)
	}

	// Instantiate in goroutine
	go func() {
		config := wazero.NewModuleConfig().WithName(uniqueName)
		_, err := h.runtime.InstantiateModule(ctx, compiled, config)
		if err != nil {
			slog.Error("Failed to instantiate module", "unique_name", uniqueName, "error", err)
			h.collector.Counter("module_load_failure", 1, map[string]string{"service_id": serviceID, "version": version})
		} else {
			h.collector.Counter("module_load_success", 1, map[string]string{"service_id": serviceID, "version": version})
		}
	}()

	return nil
}

// SetActiveVersion updates the routing to point to a specific version.
func (h *WazeroRuntimeHost) SetActiveVersion(ctx context.Context, serviceID, version string) error {
	uniqueName := fmt.Sprintf("%s-%s", serviceID, version)

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.requests[uniqueName]; !ok {
		return fmt.Errorf("version %s of service %s is not loaded", version, serviceID)
	}

	h.activeVersions[serviceID] = uniqueName
	slog.Info("Switched active version", "service_id", serviceID, "version", version, "unique_name", uniqueName)
	return nil
}

// SetShadowVersion activates a specific version as a shadow deployment.
func (h *WazeroRuntimeHost) SetShadowVersion(ctx context.Context, serviceID, version string) error {
	uniqueName := fmt.Sprintf("%s-%s", serviceID, version)

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.requests[uniqueName]; !ok {
		return fmt.Errorf("version %s of service %s is not loaded", version, serviceID)
	}

	h.shadowVersions[serviceID] = uniqueName
	slog.Info("Switched shadow version", "service_id", serviceID, "version", version, "unique_name", uniqueName)
	return nil
}

// UnsetShadowVersion removes the shadow deployment for a service without unloading the module.
func (h *WazeroRuntimeHost) UnsetShadowVersion(ctx context.Context, serviceID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.shadowVersions[serviceID]; ok {
		delete(h.shadowVersions, serviceID)
		slog.Info("Unset shadow version", "service_id", serviceID)
	}
	return nil
}

// Invoke calls a function on the loaded module of the active version.
func (h *WazeroRuntimeHost) Invoke(ctx context.Context, serviceID, method string, payload []byte) ([]byte, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		h.collector.Histogram("invoke_duration_seconds", duration, map[string]string{"service_id": serviceID, "method": method})
	}()

	h.mu.RLock()
	uniqueName, active := h.activeVersions[serviceID]
	shadowName, shadow := h.shadowVersions[serviceID]

	if !active {
		h.mu.RUnlock()
		h.collector.Counter("invoke_error", 1, map[string]string{"service_id": serviceID, "type": "no_active_version"})
		return nil, fmt.Errorf("service %s has no active version", serviceID)
	}

	reqCh, exists := h.requests[uniqueName]
	var shadowReqCh chan request
	if shadow {
		shadowReqCh = h.requests[shadowName]
	}
	h.mu.RUnlock()

	if !exists {
		h.collector.Counter("invoke_error", 1, map[string]string{"service_id": serviceID, "type": "not_found"})
		return nil, fmt.Errorf("module %s not found or not ready", uniqueName)
	}

	// Shadow Invocation
	if shadow && shadowReqCh != nil {
		go func() {
			shadowCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			sResponseCh := make(chan response, 1)
			// Clone payload to avoid data race
			payloadCopy := make([]byte, len(payload))
			copy(payloadCopy, payload)

			sReq := request{
				method:     method,
				payload:    payloadCopy,
				responseCh: sResponseCh,
			}

			select {
			case shadowReqCh <- sReq:
				// Sent
			case <-shadowCtx.Done():
				h.collector.Counter("invoke_error", 1, map[string]string{"service_id": serviceID, "type": "shadow_timeout_send"})
				return
			}

			select {
			case res := <-sResponseCh:
				if res.err != nil {
					h.collector.Counter("invoke_error", 1, map[string]string{"service_id": serviceID, "type": "shadow_execution_error"})
				} else {
					h.collector.Counter("invoke_success", 1, map[string]string{"service_id": serviceID, "type": "shadow"})
				}
			case <-shadowCtx.Done():
				h.collector.Counter("invoke_error", 1, map[string]string{"service_id": serviceID, "type": "shadow_timeout_recv"})
			}
		}()
	}

	responseCh := make(chan response, 1) // Buffered
	req := request{
		method:     method,
		payload:    payload,
		responseCh: responseCh,
	}

	select {
	case reqCh <- req:
		// Request sent
	case <-ctx.Done():
		h.collector.Counter("invoke_error", 1, map[string]string{"service_id": serviceID, "type": "context_cancelled"})
		return nil, ctx.Err()
	}

	select {
	case res := <-responseCh:
		if res.err != nil {
			h.collector.Counter("invoke_error", 1, map[string]string{"service_id": serviceID, "type": "execution_error"})
		} else {
			h.collector.Counter("invoke_success", 1, map[string]string{"service_id": serviceID})
		}
		return res.payload, res.err
	case <-ctx.Done():
		h.collector.Counter("invoke_error", 1, map[string]string{"service_id": serviceID, "type": "timeout"})
		return nil, ctx.Err()
	}
}

// UnloadVersion removes a specific version of a module.
func (h *WazeroRuntimeHost) UnloadVersion(ctx context.Context, serviceID, version string) error {
	uniqueName := fmt.Sprintf("%s-%s", serviceID, version)

	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if it's the active version
	if active, ok := h.activeVersions[serviceID]; ok && active == uniqueName {
		slog.Warn("Unloading active version", "service_id", serviceID, "version", version)
		delete(h.activeVersions, serviceID)
	}

	// Check if it's the shadow version
	if shadow, ok := h.shadowVersions[serviceID]; ok && shadow == uniqueName {
		slog.Warn("Unloading shadow version", "service_id", serviceID, "version", version)
		delete(h.shadowVersions, serviceID)
	}

	// Ensure module is closed using runtime reference if possible
	if mod := h.runtime.Module(uniqueName); mod != nil {
		if err := mod.Close(ctx); err != nil {
			slog.Warn("Failed to close module via runtime", "unique_name", uniqueName, "error", err)
		}
	} else if mod, exists := h.modules[uniqueName]; exists {
		// Fallback
		mod.Close(ctx)
	}

	delete(h.modules, uniqueName)
	delete(h.requests, uniqueName)
	delete(h.currentReq, uniqueName)

	return nil
}

// Close closes the runtime.
func (h *WazeroRuntimeHost) Close(ctx context.Context) error {
	return h.runtime.Close(ctx)
}
