package runtime

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"go.opentelemetry.io/otel/trace"

	"ghost-ops/pkg/protocol"
)

// Request represents a pending invocation.
type Request struct {
	method     string
	payload    []byte
	traceID    string
	spanID     string
	responseCh chan Response
}

// Response represents the result of an invocation.
type Response struct {
	payload []byte
	err     error
}

// ThreadSafeBuffer is a goroutine-safe bytes.Buffer.
type ThreadSafeBuffer struct {
	b  bytes.Buffer
	mu sync.Mutex
}

func (b *ThreadSafeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func (b *ThreadSafeBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Return a copy to avoid race conditions if the caller modifies it
	out := make([]byte, b.b.Len())
	copy(out, b.b.Bytes())
	return out
}

func (b *ThreadSafeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}

// WazeroRuntimeHost implements RuntimeHost using wazero.
type WazeroRuntimeHost struct {
	runtime        wazero.Runtime
	modules        map[string]api.Module        // Key: uniqueName (serviceID-version)
	requests       map[string]chan Request      // Key: uniqueName
	currentReq     map[string]Request           // Key: uniqueName
	activeVersions map[string]string            // Key: serviceID, Value: uniqueName
	shadowVersions map[string]string            // Key: serviceID, Value: uniqueName
	logBuffers     map[string]*ThreadSafeBuffer // Key: uniqueName

	// Compiled Module Cache
	compiledCache map[string]wazero.CompiledModule // Key: hash of WASM bytes
	cacheOrder    []string                         // LRU order (oldest first)

	mu        sync.RWMutex
	store     protocol.StateStore
	collector protocol.MetricsCollector
}

const maxCacheSize = 50

// NewWazeroRuntimeHost creates a new WazeroRuntimeHost.
func NewWazeroRuntimeHost(ctx context.Context, store protocol.StateStore, collector protocol.MetricsCollector) (*WazeroRuntimeHost, error) {
	// 128MB limit per memory instance
	// WithCloseOnContextDone ensures that long-running or infinite loops are terminated when context is cancelled.
	config := wazero.NewRuntimeConfig().
		WithMemoryLimitPages(2048).
		WithCloseOnContextDone(true)

	r := wazero.NewRuntimeWithConfig(ctx, config)

	// Instantiate WASI, as many modules might need it.
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	h := &WazeroRuntimeHost{
		runtime:        r,
		modules:        make(map[string]api.Module),
		requests:       make(map[string]chan Request),
		currentReq:     make(map[string]Request),
		activeVersions: make(map[string]string),
		shadowVersions: make(map[string]string),
		logBuffers:     make(map[string]*ThreadSafeBuffer),
		compiledCache:  make(map[string]wazero.CompiledModule),
		cacheOrder:     make([]string, 0, maxCacheSize),
		store:          store,
		collector:      collector,
	}

	// Instantiate host module
	_, err := r.NewHostModuleBuilder("ghost_ops").
		NewFunctionBuilder().WithFunc(h.kvGet).Export("kv_get").
		NewFunctionBuilder().WithFunc(h.rpc).Export("rpc").
		NewFunctionBuilder().WithFunc(h.nextCommand).Export("next_command").
		NewFunctionBuilder().WithFunc(h.readPayload).Export("read_payload").
		NewFunctionBuilder().WithFunc(h.submitResult).Export("submit_result").
		NewFunctionBuilder().WithFunc(h.getTraceContext).Export("get_trace_context").
		Instantiate(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to instantiate host module: %w", err)
	}

	return h, nil
}

func (h *WazeroRuntimeHost) kvGet(ctx context.Context, m api.Module, keyPtr, keyLen, valPtr, valLen uint32) uint32 {
	keyBytes, ok := m.Memory().Read(keyPtr, keyLen)
	if !ok {
		return 0
	}
	key := string(keyBytes)

	val, err := h.store.Get(ctx, key)
	if err != nil {
		return 0
	}

	if uint64(len(val)) > uint64(valLen) {
		if !m.Memory().Write(valPtr, val[:valLen]) {
			return 0
		}
	} else {
		if !m.Memory().Write(valPtr, val) {
			return 0
		}
	}

	l := len(val)
	if l > math.MaxUint32 {
		l = math.MaxUint32
	}
	return uint32(l)
}

func (h *WazeroRuntimeHost) rpc(ctx context.Context, m api.Module, svcPtr, svcLen, methPtr, methLen, payPtr, payLen, outPtr, outLen uint32) uint32 {
	mem := m.Memory()
	svcBytes, ok := mem.Read(svcPtr, svcLen)
	if !ok {
		return 0
	}
	targetServiceID := string(svcBytes)

	methBytes, ok := mem.Read(methPtr, methLen)
	if !ok {
		return 0
	}
	method := string(methBytes)

	payload, ok := mem.Read(payPtr, payLen)
	if !ok {
		return 0
	}
	payloadCopy := make([]byte, len(payload))
	copy(payloadCopy, payload)

	result, err := h.Invoke(ctx, targetServiceID, method, payloadCopy)
	if err != nil {
		return 0
	}

	if uint64(len(result)) > uint64(outLen) {
		if !mem.Write(outPtr, result[:outLen]) {
			return 0
		}
	} else {
		if !mem.Write(outPtr, result) {
			return 0
		}
	}

	l := len(result)
	if l > math.MaxUint32 {
		l = math.MaxUint32
	}
	return uint32(l)
}

func (h *WazeroRuntimeHost) nextCommand(ctx context.Context, m api.Module, methPtr, methCap uint32) uint64 {
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
		if uint64(len(methBytes)) > uint64(methCap) {
			methBytes = methBytes[:methCap]
		}
		m.Memory().Write(methPtr, methBytes)

		payLen := len(req.payload)
		if payLen > math.MaxUint32 {
			payLen = math.MaxUint32
		}
		methLen := len(methBytes)
		if methLen > math.MaxUint32 {
			methLen = math.MaxUint32
		}
		return (uint64(payLen) << 32) | uint64(methLen)

	case <-ctx.Done():
		return 0
	}
}

func (h *WazeroRuntimeHost) readPayload(ctx context.Context, m api.Module, ptr, cap uint32) {
	moduleName := m.Name()
	h.mu.RLock()
	req, ok := h.currentReq[moduleName]
	h.mu.RUnlock()
	if !ok {
		return
	}

	if uint64(len(req.payload)) > uint64(cap) {
		m.Memory().Write(ptr, req.payload[:cap])
	} else {
		m.Memory().Write(ptr, req.payload)
	}
}

func (h *WazeroRuntimeHost) submitResult(ctx context.Context, m api.Module, ptr, len uint32) {
	moduleName := m.Name()
	h.mu.RLock()
	req, ok := h.currentReq[moduleName]
	h.mu.RUnlock()
	if !ok {
		return
	}

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
	case req.responseCh <- Response{payload: res}:
	default:
	}
}

func (h *WazeroRuntimeHost) getTraceContext(ctx context.Context, m api.Module, ptr, cap uint32) uint32 {
	moduleName := m.Name()
	h.mu.RLock()
	req, ok := h.currentReq[moduleName]
	h.mu.RUnlock()
	if !ok {
		return 0
	}

	var val string
	if req.traceID != "" {
		val = fmt.Sprintf("trace_id=%s;span_id=%s", req.traceID, req.spanID)
	}

	if len(val) == 0 {
		return 0
	}

	if uint64(len(val)) > uint64(cap) {
		if !m.Memory().Write(ptr, []byte(val)[:cap]) {
			return 0
		}
	} else {
		if !m.Memory().Write(ptr, []byte(val)) {
			return 0
		}
	}

	l := len(val)
	if l > math.MaxUint32 {
		l = math.MaxUint32
	}
	return uint32(l)
}

// LoadModule loads a WASM binary for a specific version.
func (h *WazeroRuntimeHost) LoadModule(ctx context.Context, serviceID, version string, wasmBytes []byte) error {
	uniqueName := fmt.Sprintf("%s-%s", serviceID, version)

	// Compute hash for cache key
	hash := sha256.Sum256(wasmBytes)
	hashKey := hex.EncodeToString(hash[:])

	h.mu.Lock()
	if _, exists := h.modules[uniqueName]; exists {
		h.mu.Unlock()
		return protocol.NewAppError(protocol.ErrConflict.Code(), fmt.Sprintf("module %s already loaded", uniqueName), nil)
	}
	// Also check if channel exists
	if _, exists := h.requests[uniqueName]; exists {
		h.mu.Unlock()
		return protocol.NewAppError(protocol.ErrConflict.Code(), fmt.Sprintf("request channel for %s already exists", uniqueName), nil)
	}

	h.requests[uniqueName] = make(chan Request)

	// Create log buffer
	logBuf := &ThreadSafeBuffer{}
	h.logBuffers[uniqueName] = logBuf

	// Check Cache
	compiled, cached := h.compiledCache[hashKey]
	if cached {
		// Move to end of LRU list (mark as recently used)
		for i, k := range h.cacheOrder {
			if k == hashKey {
				// Remove from current position
				h.cacheOrder = append(h.cacheOrder[:i], h.cacheOrder[i+1:]...)
				break
			}
		}
		h.cacheOrder = append(h.cacheOrder, hashKey)
		h.collector.Counter("module_cache_hit", 1, map[string]string{"service_id": serviceID})
	}

	h.mu.Unlock()

	if !cached {
		var err error
		compiled, err = h.runtime.CompileModule(ctx, wasmBytes)
		if err != nil {
			h.mu.Lock()
			delete(h.requests, uniqueName)
			delete(h.logBuffers, uniqueName)
			h.mu.Unlock()
			return protocol.NewAppError(protocol.ErrInternal.Code(), "failed to compile module", err)
		}

		h.mu.Lock()
		// Add to cache
		if _, exists := h.compiledCache[hashKey]; !exists {
			if len(h.compiledCache) >= maxCacheSize {
				// Evict oldest
				oldestKey := h.cacheOrder[0]
				h.cacheOrder = h.cacheOrder[1:]
				// Do NOT call Close() on the evicted module because active instances might still
				// be using it (or rather, we don't track references to know if it's safe).
				// Wazero Runtime.Close() will eventually clean up all compiled modules.
				delete(h.compiledCache, oldestKey)
			}
			h.compiledCache[hashKey] = compiled
			h.cacheOrder = append(h.cacheOrder, hashKey)
		}
		h.collector.Counter("module_cache_miss", 1, map[string]string{"service_id": serviceID})
		h.mu.Unlock()
	}

	// Instantiate in goroutine
	// Use a background context to ensure instantiation survives the request context.
	// The runtime.Close() will handle cleanup.
	asyncCtx := context.Background()
	go func() {
		config := wazero.NewModuleConfig().
			WithName(uniqueName).
			WithStdout(logBuf).
			WithStderr(logBuf)

		_, err := h.runtime.InstantiateModule(asyncCtx, compiled, config)
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
		return protocol.NewAppError(protocol.ErrNotFound.Code(), fmt.Sprintf("version %s of service %s is not loaded", version, serviceID), nil)
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
		return protocol.NewAppError(protocol.ErrNotFound.Code(), fmt.Sprintf("version %s of service %s is not loaded", version, serviceID), nil)
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
	// Enforce 5s timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

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
		return nil, protocol.NewAppError(protocol.ErrNotFound.Code(), fmt.Sprintf("service %s has no active version", serviceID), nil)
	}

	reqCh, exists := h.requests[uniqueName]
	var shadowReqCh chan Request
	if shadow {
		shadowReqCh = h.requests[shadowName]
	}
	h.mu.RUnlock()

	if !exists {
		h.collector.Counter("invoke_error", 1, map[string]string{"service_id": serviceID, "type": "not_found"})
		return nil, protocol.NewAppError(protocol.ErrNotFound.Code(), fmt.Sprintf("module %s not found or not ready", uniqueName), nil)
	}

	// Capture Trace Context
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	var traceID, spanID string
	if spanCtx.IsValid() {
		traceID = spanCtx.TraceID().String()
		spanID = spanCtx.SpanID().String()
	}

	// Shadow Invocation
	if shadow && shadowReqCh != nil {
		go func() {
			shadowCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			sResponseCh := make(chan Response, 1)
			// Clone payload to avoid data race
			payloadCopy := make([]byte, len(payload))
			copy(payloadCopy, payload)

			sReq := Request{
				method:     method,
				payload:    payloadCopy,
				traceID:    traceID,
				spanID:     spanID,
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

	responseCh := make(chan Response, 1) // Buffered
	req := Request{
		method:     method,
		payload:    payload,
		traceID:    traceID,
		spanID:     spanID,
		responseCh: responseCh,
	}

	select {
	case reqCh <- req:
		// Request sent
	case <-ctx.Done():
		h.collector.Counter("invoke_error", 1, map[string]string{"service_id": serviceID, "type": "context_cancelled"})
		return nil, protocol.NewAppError(protocol.ErrTimeout.Code(), "request context cancelled", ctx.Err())
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
		return nil, protocol.NewAppError(protocol.ErrTimeout.Code(), "operation timed out", ctx.Err())
	}
}

// CheckHealth checks if a service is healthy (loaded and responsive).
func (h *WazeroRuntimeHost) CheckHealth(ctx context.Context, serviceID string) error {
	// Check context first
	if err := ctx.Err(); err != nil {
		return protocol.NewAppError(protocol.ErrInternal.Code(), "health check context cancelled", err)
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	uniqueName, active := h.activeVersions[serviceID]
	if !active {
		return protocol.NewAppError(protocol.ErrNotFound.Code(), fmt.Sprintf("service %s has no active version", serviceID), nil)
	}

	if _, exists := h.modules[uniqueName]; !exists {
		return protocol.NewAppError(protocol.ErrInternal.Code(), fmt.Sprintf("module %s not loaded in runtime", uniqueName), nil)
	}

	// Ensure request channel exists and is not closed (though we can't easily check closed state without reading)
	// We check if it exists in the map.
	if _, exists := h.requests[uniqueName]; !exists {
		return protocol.NewAppError(protocol.ErrInternal.Code(), fmt.Sprintf("request channel for %s missing", uniqueName), nil)
	}

	return nil
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
	delete(h.logBuffers, uniqueName)

	return nil
}

// GetActiveServiceCount returns the number of active services.
func (h *WazeroRuntimeHost) GetActiveServiceCount(ctx context.Context) (int, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.activeVersions), nil
}

// GetLogs retrieves the logs for a specific service.
func (h *WazeroRuntimeHost) GetLogs(ctx context.Context, serviceID string) ([]byte, error) {
	h.mu.RLock()
	uniqueName, active := h.activeVersions[serviceID]
	h.mu.RUnlock()

	if !active {
		return nil, protocol.NewAppError(protocol.ErrNotFound.Code(), fmt.Sprintf("service %s has no active version", serviceID), nil)
	}

	h.mu.RLock()
	logBuf, exists := h.logBuffers[uniqueName]
	h.mu.RUnlock()

	if !exists {
		// Should not happen if loaded correctly, but handle gracefully
		return []byte{}, nil
	}

	return logBuf.Bytes(), nil
}

// Close closes the runtime.
func (h *WazeroRuntimeHost) Close(ctx context.Context) error {
	return h.runtime.Close(ctx)
}
