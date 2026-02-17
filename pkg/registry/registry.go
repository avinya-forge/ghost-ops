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
	wasmBytes, err := r.engine.Evolve(ctx, *bp)
	if err != nil {
		slog.Error("Failed to evolve service", "service_id", bp.ServiceID, "error", err)
		r.collector.Counter("reconcile_failure", 1, map[string]string{"phase": "evolve", "service_id": bp.ServiceID})
		return true, nil // Return true to continue processing next blueprint even if this one failed
	}

	// Calculate hash
	hash := sha256.Sum256(wasmBytes)
	hashStr := hex.EncodeToString(hash[:])

	// Check if service already exists
	existing, err := r.store.GetService(ctx, bp.ServiceID)
	if err != nil {
		slog.Error("Failed to get service state", "service_id", bp.ServiceID, "error", err)
		return true, nil
	}

	var newVersion = 1
	if existing != nil {
		newVersion = existing.Version + 1
		// Unload if exists
		slog.Info("Service exists, unloading first", "service_id", bp.ServiceID, "version", existing.Version)
		if err := r.runtime.UnloadModule(ctx, bp.ServiceID); err != nil {
			slog.Warn("Failed to unload module (might not be loaded)", "service_id", bp.ServiceID, "error", err)
		}
	}

	// Load Module
	if err := r.runtime.LoadModule(ctx, bp.ServiceID, wasmBytes); err != nil {
		slog.Error("Failed to load module", "service_id", bp.ServiceID, "error", err)
		r.collector.Counter("reconcile_failure", 1, map[string]string{"phase": "load_module", "service_id": bp.ServiceID})
		return true, nil
	}

	// Update Store
	record := protocol.ServiceRecord{
		ServiceID:          bp.ServiceID,
		Version:            newVersion,
		WASMHash:           hashStr,
		CurrentState:       protocol.StateActive,
		SynthesisTimestamp: time.Now().UTC(),
	}

	if err := r.store.UpdateService(ctx, record); err != nil {
		slog.Error("Failed to update store", "service_id", bp.ServiceID, "error", err)
		// Should we unload if store update fails? Probably not for MVP.
		r.collector.Counter("reconcile_failure", 1, map[string]string{"phase": "update_store", "service_id": bp.ServiceID})
		return true, nil
	}

	slog.Info("Service reconciled successfully", "service_id", bp.ServiceID)
	r.collector.Counter("reconcile_success", 1, map[string]string{"service_id": bp.ServiceID})

	// Update active services gauge
	services, err := r.store.ListServices(ctx)
	if err == nil {
		r.collector.Gauge("active_services", float64(len(services)), nil)
	}

	return true, nil
}

// ListServices returns all services from the store.
func (r *Registry) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	return r.store.ListServices(ctx)
}
