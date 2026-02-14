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
	store   protocol.StateStore
	engine  protocol.EvolutionEngine
	source  protocol.IntentSource
	runtime protocol.RuntimeHost
}

// NewRegistry creates a new Registry.
func NewRegistry(
	store protocol.StateStore,
	engine protocol.EvolutionEngine,
	source protocol.IntentSource,
	runtime protocol.RuntimeHost,
) *Registry {
	return &Registry{
		store:   store,
		engine:  engine,
		source:  source,
		runtime: runtime,
	}
}

// Reconcile checks for new intents and updates services.
func (r *Registry) Reconcile(ctx context.Context) error {
	slog.Info("Starting reconciliation")

	count := 0
	for {
		// Get next blueprint
		bp, err := r.source.GetNextBlueprint(ctx)
		if err != nil {
			return fmt.Errorf("failed to get blueprint: %w", err)
		}
		if bp == nil {
			break // No more blueprints
		}

		slog.Info("Processing blueprint", "service_id", bp.ServiceID)

		// Evolve
		wasmBytes, err := r.engine.Evolve(ctx, *bp)
		if err != nil {
			slog.Error("Failed to evolve service", "service_id", bp.ServiceID, "error", err)
			continue
		}

		// Calculate hash
		hash := sha256.Sum256(wasmBytes)
		hashStr := hex.EncodeToString(hash[:])

		// Check if service already exists
		existing, err := r.store.GetService(ctx, bp.ServiceID)
		if err != nil {
			slog.Error("Failed to get service state", "service_id", bp.ServiceID, "error", err)
			continue
		}

		if existing != nil {
			// Unload if exists
			slog.Info("Service exists, unloading first", "service_id", bp.ServiceID)
			if err := r.runtime.UnloadModule(ctx, bp.ServiceID); err != nil {
				slog.Warn("Failed to unload module (might not be loaded)", "service_id", bp.ServiceID, "error", err)
			}
		}

		// Load Module
		if err := r.runtime.LoadModule(ctx, bp.ServiceID, wasmBytes); err != nil {
			slog.Error("Failed to load module", "service_id", bp.ServiceID, "error", err)
			continue
		}

		// Update Store
		record := protocol.ServiceRecord{
			ServiceID:          bp.ServiceID,
			WASMHash:           hashStr,
			CurrentState:       protocol.StateActive,
			SynthesisTimestamp: time.Now().UTC(),
		}

		if err := r.store.UpdateService(ctx, record); err != nil {
			slog.Error("Failed to update store", "service_id", bp.ServiceID, "error", err)
			// Should we unload if store update fails? Probably not for MVP.
			continue
		}

		slog.Info("Service reconciled successfully", "service_id", bp.ServiceID)
		count++
	}

	slog.Info("Reconciliation finished", "processed_count", count)
	return nil
}

// ListServices returns all services from the store.
func (r *Registry) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	return r.store.ListServices(ctx)
}
