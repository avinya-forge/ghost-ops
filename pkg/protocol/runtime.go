package protocol

import "context"

// RuntimeHost executes the evolved WASM code.
type RuntimeHost interface {
	// LoadModule loads a WASM binary.
	LoadModule(ctx context.Context, serviceID string, wasmBytes []byte) error

	// Invoke calls a function on the loaded module.
	Invoke(ctx context.Context, serviceID, method string, payload []byte) ([]byte, error)

	// UnloadModule removes a module.
	UnloadModule(ctx context.Context, serviceID string) error
}
