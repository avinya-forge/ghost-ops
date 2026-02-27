//go:build !wasip1

package guest

import "fmt"

// Handler is a function that processes a request.
type Handler func(payload []byte) ([]byte, error)

// Register registers a handler for a method.
func Register(method string, h Handler) {
	// No-op on host
}

// Start begins the event loop. This function never returns.
func Start() {
	// Block forever on host to simulate
	select {}
}

// Get retrieves a value from the state store.
func Get(key string) ([]byte, error) {
	return nil, fmt.Errorf("guest SDK functions are only available in WASM/WASI environment")
}

// Call invokes a method on another service.
func Call(service, method string, payload []byte) ([]byte, error) {
	return nil, fmt.Errorf("guest SDK functions are only available in WASM/WASI environment")
}

// GetTraceContext retrieves the current trace context as a map.
func GetTraceContext() (map[string]string, error) {
	return nil, fmt.Errorf("guest SDK functions are only available in WASM/WASI environment")
}
