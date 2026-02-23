package protocol

import "context"

// RuntimeHost executes the evolved WASM code.
type RuntimeHost interface {
	// LoadModule loads a WASM binary for a specific version.
	LoadModule(ctx context.Context, serviceID, version string, wasmBytes []byte) error

	// Invoke calls a function on the loaded module of the active version.
	Invoke(ctx context.Context, serviceID, method string, payload []byte) ([]byte, error)

	// SetActiveVersion updates the routing to point to a specific version.
	SetActiveVersion(ctx context.Context, serviceID, version string) error

	// SetShadowVersion activates a specific version as a shadow deployment.
	SetShadowVersion(ctx context.Context, serviceID, version string) error

	// UnsetShadowVersion removes the shadow deployment for a service without unloading the module.
	UnsetShadowVersion(ctx context.Context, serviceID string) error

	// UnloadVersion removes a specific version of a module.
	UnloadVersion(ctx context.Context, serviceID, version string) error

	// CheckHealth checks if a service is healthy (loaded and responsive).
	CheckHealth(ctx context.Context, serviceID string) error

	// GetLogs retrieves the logs for a specific service.
	GetLogs(ctx context.Context, serviceID string) ([]byte, error)

	// Close closes the runtime.
	Close(ctx context.Context) error
}
