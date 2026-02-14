package registry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"ghost-ops/pkg/protocol"
)

// ServiceRegistry orchestrates the lifecycle of services.
type ServiceRegistry struct {
	intentSource    protocol.IntentSource
	evolutionEngine protocol.EvolutionEngine
	runtimeHost     protocol.RuntimeHost
	stateStore      protocol.StateStore
}

// NewServiceRegistry creates a new ServiceRegistry.
func NewServiceRegistry(
	intent protocol.IntentSource,
	evolution protocol.EvolutionEngine,
	runtime protocol.RuntimeHost,
	store protocol.StateStore,
) *ServiceRegistry {
	return &ServiceRegistry{
		intentSource:    intent,
		evolutionEngine: evolution,
		runtimeHost:     runtime,
		stateStore:      store,
	}
}

// Reconcile processes one blueprint from the intent source.
// It returns true if a blueprint was processed, false if no more blueprints are available.
func (r *ServiceRegistry) Reconcile(ctx context.Context) (bool, error) {
	// 1. Get Intent
	bp, err := r.intentSource.GetNextBlueprint(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get blueprint: %w", err)
	}
	if bp == nil {
		return false, nil // No more blueprints
	}

	// 2. Evolve (compile/generate WASM)
	wasmBytes, err := r.evolutionEngine.Evolve(ctx, *bp)
	if err != nil {
		return true, fmt.Errorf("failed to evolve service %s: %w", bp.ServiceID, err)
	}

	// Calculate hash
	hash := sha256.Sum256(wasmBytes)
	hashStr := hex.EncodeToString(hash[:])

	// 3. Update Runtime
	if err := r.runtimeHost.LoadModule(ctx, bp.ServiceID, wasmBytes); err != nil {
		return true, fmt.Errorf("failed to load module for service %s: %w", bp.ServiceID, err)
	}

	// 4. Update State Store
	record := protocol.ServiceRecord{
		ServiceID:          bp.ServiceID,
		WASMHash:           hashStr,
		CurrentState:       protocol.StateActive,
		SynthesisTimestamp: time.Now(),
	}

	if err := r.stateStore.UpdateService(ctx, record); err != nil {
		return true, fmt.Errorf("failed to update state store for service %s: %w", bp.ServiceID, err)
	}

	return true, nil
}

// GetServiceStatus returns the status of a service from the store.
func (r *ServiceRegistry) GetServiceStatus(ctx context.Context, serviceID string) (*protocol.ServiceRecord, error) {
	return r.stateStore.GetService(ctx, serviceID)
}

// ListServices returns all services from the store.
func (r *ServiceRegistry) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	return r.stateStore.ListServices(ctx)
}
