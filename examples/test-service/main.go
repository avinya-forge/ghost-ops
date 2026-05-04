// Used by pkg/api/integration_test.go and pkg/runtime/wazero_host_async_verify_test.go
// to compile a real WASM module against the guest SDK. Do not delete.
package main

import (
	"ghost-ops/pkg/sdk/guest"
)

func main() {
	guest.Register("echo", func(payload []byte) ([]byte, error) {
		return payload, nil
	})
	guest.Start()
}
