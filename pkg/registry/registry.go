package registry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"ghost-ops/pkg/config"
	"ghost-ops/pkg/logging"
	"ghost-ops/pkg/protocol"
)

// Registry orchestrates the lifecycle of services.
type Registry struct {
	store     protocol.StateStore
	engine    protocol.EvolutionEngine
	source    protocol.IntentSource
	runtime   protocol.RuntimeHost
	collector protocol.MetricsCollector
	eventBus  protocol.EventBus
	config    config.RegistryConfig
}

// NewRegistry creates a new Registry.
func NewRegistry(
	store protocol.StateStore,
	engine protocol.EvolutionEngine,
	source protocol.IntentSource,
	runtime protocol.RuntimeHost,
	collector protocol.MetricsCollector,
	eventBus protocol.EventBus,
	cfg config.RegistryConfig,
) *Registry {
	return &Registry{
		store:     store,
		engine:    engine,
		source:    source,
		runtime:   runtime,
		collector: collector,
		eventBus:  eventBus,
		config:    cfg,
	}
}

// Reconcile processes a single pending intent.
// It returns true if a blueprint was processed, false if no more blueprints are available.
func (r *Registry) Reconcile(ctx context.Context) (bool, error) {
	slog.Debug("Checking for next blueprint")

	// Get next blueprint
	bp, err := r.source.GetNextBlueprint(ctx)
	if err != nil {
		return false, wrapError(protocol.ErrInternal.Code(), "failed to get blueprint", err)
	}
	if bp == nil {
		slog.Debug("No more blueprints to process")
		return false, nil // No more blueprints
	}

	slog.Info("Processing blueprint", "service_id", bp.ServiceID)

	start := time.Now()
	defer func() {
		r.collector.Histogram("reconcile_duration_seconds", time.Since(start).Seconds(), map[string]string{"service_id": bp.ServiceID})
	}()

	// Dependency Check
	allServices, err := r.store.ListServices(ctx)
	if err != nil {
		slog.Error("Failed to list services for dependency check", "error", err)
		r.collector.Counter("reconcile_failure", 1, map[string]string{"phase": "dependency_check", "service_id": bp.ServiceID})
		return true, nil
	}

	graph := NewServiceGraph(allServices)
	if err := graph.AddService(bp.ServiceID, bp.Dependencies); err != nil {
		slog.Error("Blueprint rejected due to dependency cycle", "service_id", bp.ServiceID, "error", err)
		r.collector.Counter("reconcile_failure", 1, map[string]string{"phase": "validation", "service_id": bp.ServiceID})
		return true, nil
	}

	// Check if service already exists
	existing, err := r.store.GetService(ctx, bp.ServiceID)
	if err != nil {
		slog.Error("Failed to get service state", "service_id", bp.ServiceID, "error", err)
		return true, nil
	}

	// Resource Limit Check for New Services
	if existing == nil || existing.ActiveVersion == 0 {
		count, err := r.runtime.GetActiveServiceCount(ctx)
		if err != nil {
			slog.Error("Failed to get active service count", "error", err)
			r.collector.Counter("reconcile_failure", 1, map[string]string{"phase": "resource_check", "service_id": bp.ServiceID})
			return true, nil
		}
		if r.config.MaxActiveServices > 0 && count >= r.config.MaxActiveServices {
			slog.Warn("Max active services limit reached, skipping blueprint", "limit", r.config.MaxActiveServices, "current", count, "service_id", bp.ServiceID)
			r.collector.Counter("reconcile_failure", 1, map[string]string{"phase": "resource_limit", "service_id": bp.ServiceID})
			return true, nil
		}
	}

	// Evolve
	evolveStart := time.Now()
	wasmBytes, hashStr, err := r.evolveService(ctx, bp)
	r.collector.Histogram("evolve_duration_seconds", time.Since(evolveStart).Seconds(), map[string]string{"service_id": bp.ServiceID})

	if err != nil {
		slog.Error("Failed to evolve service", "service_id", bp.ServiceID, "error", err)
		r.collector.Counter("reconcile_failure", 1, map[string]string{"phase": "evolve", "service_id": bp.ServiceID})
		return true, nil // Return true to continue processing next blueprint even if this one failed
	}

	// Calculate new version
	var newVersion = 1
	if existing != nil {
		newVersion = existing.Version + 1
	}

	// Deploy to Runtime
	deployStart := time.Now()
	err = r.deployToRuntime(ctx, bp.ServiceID, existing, newVersion, wasmBytes, bp.Constraints)
	r.collector.Histogram("deploy_duration_seconds", time.Since(deployStart).Seconds(), map[string]string{"service_id": bp.ServiceID})

	if err != nil {
		slog.Error("Failed to load module", "service_id", bp.ServiceID, "error", err)
		r.collector.Counter("reconcile_failure", 1, map[string]string{"phase": "load_module", "service_id": bp.ServiceID})
		return true, nil
	}

	// Update Store
	if err := r.updateStore(ctx, bp.ServiceID, newVersion, hashStr, existing, bp.Constraints, bp.Dependencies); err != nil {
		slog.Error("Failed to update store", "service_id", bp.ServiceID, "error", err)
		r.collector.Counter("reconcile_failure", 1, map[string]string{"phase": "update_store", "service_id": bp.ServiceID})
		return true, nil
	}

	slog.Info("Service reconciled successfully", "service_id", bp.ServiceID)
	r.collector.Counter("reconcile_success", 1, map[string]string{"service_id": bp.ServiceID})

	// Publish Event
	if r.eventBus != nil {
		r.eventBus.Publish(ctx, protocol.Event{
			Type:      protocol.EventServiceDeployed,
			ServiceID: bp.ServiceID,
			Payload: map[string]interface{}{
				"version": newVersion,
			},
		})
	}

	// Refresh Metrics
	r.refreshMetrics(ctx)

	return true, nil
}

// StartEventLoop starts listening for internal system events (like promotion/rollback).
func (r *Registry) StartEventLoop(ctx context.Context) {
	if r.eventBus == nil {
		return
	}

	promoteCh, err := r.eventBus.Subscribe(ctx, protocol.EventPromotionRequired)
	if err != nil {
		slog.Error("Failed to subscribe to promotion events", "error", err)
	}

	rollbackCh, err := r.eventBus.Subscribe(ctx, protocol.EventRollbackRequired)
	if err != nil {
		slog.Error("Failed to subscribe to rollback events", "error", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-promoteCh:
				slog.Info("Promotion event received", "service_id", evt.ServiceID, "reason", evt.Payload["reason"])
				if err := r.PromoteService(ctx, evt.ServiceID); err != nil {
					slog.Error("Failed to promote service", "service_id", evt.ServiceID, "error", err)
				}
			case evt := <-rollbackCh:
				slog.Info("Rollback event received", "service_id", evt.ServiceID, "reason", evt.Payload["reason"])
				if err := r.RollbackService(ctx, evt.ServiceID); err != nil {
					slog.Error("Failed to rollback service", "service_id", evt.ServiceID, "error", err)
				}
			}
		}
	}()
}

// StartHealthCheck starts a background routine to check service health.
func (r *Registry) StartHealthCheck(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.CheckServices(ctx)
			}
		}
	}()
}

// CheckServices performs a health check on all services.
func (r *Registry) CheckServices(ctx context.Context) {
	services, err := r.store.ListServices(ctx)
	if err != nil {
		slog.Error("Failed to list services for health check", "error", err)
		return
	}

	for _, svc := range services {
		// Only check services that are supposed to be active
		if svc.ActiveVersion > 0 {
			if err := r.runtime.CheckHealth(ctx, svc.ServiceID); err != nil {
				slog.Warn("Service unhealthy", "service_id", svc.ServiceID, "error", err)
				r.collector.Counter("health_check_failure", 1, map[string]string{"service_id": svc.ServiceID})

				// Publish Event
				if r.eventBus != nil {
					r.eventBus.Publish(ctx, protocol.Event{
						Type:      protocol.EventServiceUnhealthy,
						ServiceID: svc.ServiceID,
						Payload: map[string]interface{}{
							"error": err.Error(),
						},
					})
				}

				// Purge from runtime to ensure clean state
				versionStr := fmt.Sprintf("%d", svc.ActiveVersion)
				if unloadErr := r.runtime.UnloadVersion(ctx, svc.ServiceID, versionStr); unloadErr != nil {
					slog.Error("Failed to purge unhealthy service", "service_id", svc.ServiceID, "error", unloadErr)
				}

				logging.Audit(ctx, "purge_unhealthy_service", map[string]interface{}{
					"service_id": svc.ServiceID,
					"version":    versionStr,
					"error":      err.Error(),
				})
			}
		}
	}
}

func (r *Registry) evolveService(ctx context.Context, bp *protocol.Blueprint) ([]byte, string, error) {
	wasmBytes, err := r.engine.Evolve(ctx, *bp)
	if err != nil {
		return nil, "", wrapError(protocol.ErrInternal.Code(), "failed to evolve service", err)
	}

	// Calculate hash
	hash := sha256.Sum256(wasmBytes)
	hashStr := hex.EncodeToString(hash[:])
	return wasmBytes, hashStr, nil
}

func (r *Registry) deployToRuntime(ctx context.Context, serviceID string, existing *protocol.ServiceRecord, newVersion int, wasmBytes []byte, constraints map[string]interface{}) error {
	versionStr := fmt.Sprintf("%d", newVersion)

	// Check shadow mode
	shadowMode := false
	if val, ok := constraints["shadow_mode"]; ok {
		if b, ok := val.(bool); ok && b {
			shadowMode = true
		}
	}

	// 1. Load new version
	if err := r.runtime.LoadModule(ctx, serviceID, versionStr, wasmBytes); err != nil {
		return wrapError(protocol.ErrInternal.Code(), "failed to load module", err)
	}

	if shadowMode {
		// Set Shadow Version
		if err := r.runtime.SetShadowVersion(ctx, serviceID, versionStr); err != nil {
			_ = r.runtime.UnloadVersion(ctx, serviceID, versionStr)
			return wrapError(protocol.ErrInternal.Code(), "failed to set shadow version", err)
		}

		// Unload OLD Shadow Version if exists and different
		if existing != nil && existing.ShadowVersion > 0 {
			oldShadowVer := fmt.Sprintf("%d", existing.ShadowVersion)
			if oldShadowVer != versionStr {
				_ = r.runtime.UnloadVersion(ctx, serviceID, oldShadowVer)
			}
		}
	} else {
		// 2. Set active version (Promote)
		if err := r.runtime.SetActiveVersion(ctx, serviceID, versionStr); err != nil {
			// Rollback? Unload new version
			_ = r.runtime.UnloadVersion(ctx, serviceID, versionStr)
			return wrapError(protocol.ErrInternal.Code(), "failed to set active version", err)
		}

		// 3. Unload old version if exists
		if existing != nil {
			// Unload old active
			// Fallback: use existing.Version as old active if ActiveVersion is 0 (legacy)
			oldActive := existing.ActiveVersion
			if oldActive == 0 {
				oldActive = existing.Version
			}

			if oldActive > 0 {
				oldActiveStr := fmt.Sprintf("%d", oldActive)
				if oldActiveStr != versionStr {
					slog.Info("Unloading old version", "service_id", serviceID, "old_version", oldActiveStr)
					if err := r.runtime.UnloadVersion(ctx, serviceID, oldActiveStr); err != nil {
						slog.Warn("Failed to unload old version", "service_id", serviceID, "version", oldActiveStr, "error", err)
					}
				}
			}

			// Also unload shadow if any, since we are promoting/deploying new active
			if existing.ShadowVersion > 0 {
				oldShadowVer := fmt.Sprintf("%d", existing.ShadowVersion)
				if oldShadowVer != versionStr {
					_ = r.runtime.UnloadVersion(ctx, serviceID, oldShadowVer)
				} else {
					// If the shadow version is the one being promoted, just unset it as shadow
					// The module remains loaded and is now Active (from SetActiveVersion above)
					_ = r.runtime.UnsetShadowVersion(ctx, serviceID)
				}
			}
		}
	}

	return nil
}

func (r *Registry) updateStore(ctx context.Context, serviceID string, version int, hashStr string, existing *protocol.ServiceRecord, constraints map[string]interface{}, dependencies []string) error {
	shadowMode := false
	if val, ok := constraints["shadow_mode"]; ok {
		if b, ok := val.(bool); ok && b {
			shadowMode = true
		}
	}

	record := protocol.ServiceRecord{
		ServiceID:          serviceID,
		Version:            version,
		WASMHash:           hashStr,
		SynthesisTimestamp: time.Now().UTC(),
		Dependencies:       dependencies,
	}

	if existing != nil {
		// Preserve Active fields
		record.ActiveVersion = existing.ActiveVersion
		record.ActiveWASMHash = existing.ActiveWASMHash

		// Migration for legacy records where Version was Active
		if record.ActiveVersion == 0 && existing.Version > 0 {
			record.ActiveVersion = existing.Version
			record.ActiveWASMHash = existing.WASMHash
		}
	}

	if shadowMode {
		record.ShadowVersion = version
		record.ShadowWASMHash = hashStr
		record.CurrentState = protocol.StateShadow
	} else {
		record.ActiveVersion = version
		record.ActiveWASMHash = hashStr
		record.ShadowVersion = 0
		record.ShadowWASMHash = ""
		record.CurrentState = protocol.StateActive
	}

	if err := r.store.UpdateService(ctx, record); err != nil {
		return wrapError(protocol.ErrInternal.Code(), "failed to update store", err)
	}

	logging.Audit(ctx, "update_service", map[string]interface{}{
		"service_id": serviceID,
		"version":    version,
		"hash":       hashStr,
		"state":      record.CurrentState,
	})

	return nil
}

// PromoteService promotes the current shadow version of a service to active.
func (r *Registry) PromoteService(ctx context.Context, serviceID string) error {
	existing, err := r.store.GetService(ctx, serviceID)
	if err != nil {
		return wrapError(protocol.ErrInternal.Code(), "failed to get service for promotion", err)
	}
	if existing == nil {
		return protocol.NewAppError(protocol.ErrNotFound.Code(), "service not found", nil)
	}

	if existing.ShadowVersion == 0 {
		return protocol.NewAppError(protocol.ErrInvalidInput.Code(), "no shadow version to promote", nil)
	}

	shadowVersionStr := fmt.Sprintf("%d", existing.ShadowVersion)
	oldActiveVersionStr := fmt.Sprintf("%d", existing.ActiveVersion)

	// 1. Set the shadow version as active in runtime
	if err := r.runtime.SetActiveVersion(ctx, serviceID, shadowVersionStr); err != nil {
		return wrapError(protocol.ErrInternal.Code(), "failed to promote in runtime", err)
	}

	// 2. Unload the old active version from runtime
	if existing.ActiveVersion > 0 && oldActiveVersionStr != shadowVersionStr {
		if err := r.runtime.UnloadVersion(ctx, serviceID, oldActiveVersionStr); err != nil {
			slog.Warn("Failed to unload old version during promotion", "service_id", serviceID, "version", oldActiveVersionStr, "error", err)
		}
	}

	// 3. Unset shadow in runtime (it's now active)
	_ = r.runtime.UnsetShadowVersion(ctx, serviceID)

	// 4. Update the StateStore
	existing.ActiveVersion = existing.ShadowVersion
	existing.ActiveWASMHash = existing.ShadowWASMHash
	existing.ShadowVersion = 0
	existing.ShadowWASMHash = ""
	existing.CurrentState = protocol.StateActive

	if err := r.store.UpdateService(ctx, *existing); err != nil {
		return wrapError(protocol.ErrInternal.Code(), "failed to update store during promotion", err)
	}

	logging.Audit(ctx, "promote_service", map[string]interface{}{
		"service_id": serviceID,
		"version":    existing.ActiveVersion,
	})

	if r.eventBus != nil {
		r.eventBus.Publish(ctx, protocol.Event{
			Type:      protocol.EventServiceDeployed,
			ServiceID: serviceID,
			Payload: map[string]interface{}{
				"version": existing.ActiveVersion,
				"reason":  "promotion",
			},
		})
	}

	return nil
}

// RollbackService removes the shadow version, keeping the active version intact.
func (r *Registry) RollbackService(ctx context.Context, serviceID string) error {
	existing, err := r.store.GetService(ctx, serviceID)
	if err != nil {
		return wrapError(protocol.ErrInternal.Code(), "failed to get service for rollback", err)
	}
	if existing == nil {
		return protocol.NewAppError(protocol.ErrNotFound.Code(), "service not found", nil)
	}

	if existing.ShadowVersion == 0 {
		return protocol.NewAppError(protocol.ErrInvalidInput.Code(), "no shadow version to rollback", nil)
	}

	shadowVersionStr := fmt.Sprintf("%d", existing.ShadowVersion)

	// 1. Unload shadow from runtime
	if err := r.runtime.UnloadVersion(ctx, serviceID, shadowVersionStr); err != nil {
		slog.Warn("Failed to unload shadow version during rollback", "service_id", serviceID, "version", shadowVersionStr, "error", err)
	}

	// 2. Unset shadow in runtime
	_ = r.runtime.UnsetShadowVersion(ctx, serviceID)

	// 3. Update the StateStore to clear shadow details
	rolledBackVersion := existing.ShadowVersion
	existing.ShadowVersion = 0
	existing.ShadowWASMHash = ""
	existing.CurrentState = protocol.StateActive

	if err := r.store.UpdateService(ctx, *existing); err != nil {
		return wrapError(protocol.ErrInternal.Code(), "failed to update store during rollback", err)
	}

	logging.Audit(ctx, "rollback_service", map[string]interface{}{
		"service_id":           serviceID,
		"rolled_back_version":  rolledBackVersion,
		"kept_active_version":  existing.ActiveVersion,
	})

	return nil
}

func wrapError(defaultCode string, msg string, err error) protocol.AppError {
	if appErr, ok := err.(protocol.AppError); ok {
		return protocol.NewAppError(appErr.Code(), msg, err)
	}
	return protocol.NewAppError(defaultCode, msg, err)
}

func (r *Registry) refreshMetrics(ctx context.Context) {
	services, err := r.store.ListServices(ctx)
	if err == nil {
		r.collector.Gauge("active_services", float64(len(services)), nil)
	}
}

// ListServices returns all services from the store.
func (r *Registry) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	services, err := r.store.ListServices(ctx)
	if err != nil {
		return nil, wrapError(protocol.ErrInternal.Code(), "failed to list services", err)
	}
	return services, nil
}

// GetLogs retrieves the logs for a specific service.
func (r *Registry) GetLogs(ctx context.Context, serviceID string) ([]byte, error) {
	return r.runtime.GetLogs(ctx, serviceID)
}

// Invoke calls a method on a service.
func (r *Registry) Invoke(ctx context.Context, serviceID, method string, payload []byte) ([]byte, error) {
	return r.runtime.Invoke(ctx, serviceID, method, payload)
}
