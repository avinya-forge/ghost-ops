package registry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"ghost-ops/pkg/protocol"
)

// Registry orchestrates the lifecycle of services.
type Registry struct {
	store     protocol.StateStore
	engine    protocol.EvolutionEngine
	source    protocol.IntentSource
	runtime   protocol.RuntimeHost
	collector protocol.MetricsCollector
}

// NewRegistry creates a new Registry.
func NewRegistry(
	store protocol.StateStore,
	engine protocol.EvolutionEngine,
	source protocol.IntentSource,
	runtime protocol.RuntimeHost,
	collector protocol.MetricsCollector,
) *Registry {
	return &Registry{
		store:     store,
		engine:    engine,
		source:    source,
		runtime:   runtime,
		collector: collector,
	}
}

// Reconcile processes a single pending intent.
// It returns true if a blueprint was processed, false if no more blueprints are available.
func (r *Registry) Reconcile(ctx context.Context) (bool, error) {
	slog.Debug("Checking for next blueprint")

	// Get next blueprint
	bp, err := r.source.GetNextBlueprint(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get blueprint: %w", err)
	}
	if bp == nil {
		slog.Debug("No more blueprints to process")
		return false, nil // No more blueprints
	}

	slog.Info("Processing blueprint", "service_id", bp.ServiceID)

	// Evolve
	wasmBytes, hashStr, err := r.evolveService(ctx, bp)
	if err != nil {
		slog.Error("Failed to evolve service", "service_id", bp.ServiceID, "error", err)
		r.collector.Counter("reconcile_failure", 1, map[string]string{"phase": "evolve", "service_id": bp.ServiceID})
		return true, nil // Return true to continue processing next blueprint even if this one failed
	}

	// Check if service already exists
	existing, err := r.store.GetService(ctx, bp.ServiceID)
	if err != nil {
		slog.Error("Failed to get service state", "service_id", bp.ServiceID, "error", err)
		return true, nil
	}

	// Calculate new version
	var newVersion = 1
	if existing != nil {
		newVersion = existing.Version + 1
	}

	// Deploy to Runtime
	if err := r.deployToRuntime(ctx, bp.ServiceID, existing, wasmBytes); err != nil {
		slog.Error("Failed to load module", "service_id", bp.ServiceID, "error", err)
		r.collector.Counter("reconcile_failure", 1, map[string]string{"phase": "load_module", "service_id": bp.ServiceID})
		return true, nil
	}

	// Update Store
	if err := r.updateStore(ctx, bp.ServiceID, newVersion, hashStr); err != nil {
		slog.Error("Failed to update store", "service_id", bp.ServiceID, "error", err)
		r.collector.Counter("reconcile_failure", 1, map[string]string{"phase": "update_store", "service_id": bp.ServiceID})
		return true, nil
	}

	slog.Info("Service reconciled successfully", "service_id", bp.ServiceID)
	r.collector.Counter("reconcile_success", 1, map[string]string{"service_id": bp.ServiceID})

	// Refresh Metrics
	r.refreshMetrics(ctx)

	return true, nil
}

func (r *Registry) evolveService(ctx context.Context, bp *protocol.Blueprint) ([]byte, string, error) {
	wasmBytes, err := r.engine.Evolve(ctx, *bp)
	if err != nil {
		return nil, "", err
	}

	// Calculate hash
	hash := sha256.Sum256(wasmBytes)
	hashStr := hex.EncodeToString(hash[:])
	return wasmBytes, hashStr, nil
}

func (r *Registry) deployToRuntime(ctx context.Context, serviceID string, existing *protocol.ServiceRecord, wasmBytes []byte) error {
	if existing != nil {
		// Unload if exists
		slog.Info("Service exists, unloading first", "service_id", serviceID, "version", existing.Version)
		if err := r.runtime.UnloadModule(ctx, serviceID); err != nil {
			slog.Warn("Failed to unload module (might not be loaded)", "service_id", serviceID, "error", err)
		}
	}

	return r.runtime.LoadModule(ctx, serviceID, wasmBytes)
}

func (r *Registry) updateStore(ctx context.Context, serviceID string, version int, hashStr string) error {
	record := protocol.ServiceRecord{
		ServiceID:          serviceID,
		Version:            version,
		WASMHash:           hashStr,
		CurrentState:       protocol.StateActive,
		SynthesisTimestamp: time.Now().UTC(),
	}

	return r.store.UpdateService(ctx, record)
}

func (r *Registry) refreshMetrics(ctx context.Context) {
	services, err := r.store.ListServices(ctx)
	if err == nil {
		r.collector.Gauge("active_services", float64(len(services)), nil)
	}
}

// ListServices returns all services from the store.
func (r *Registry) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	return r.store.ListServices(ctx)
}
