//go:build wasip1

package guest

import (
	"fmt"
	"unsafe"
)

// Handler is a function that processes a request.
type Handler func(payload []byte) ([]byte, error)

var handlers = make(map[string]Handler)

// Register registers a handler for a method.
func Register(method string, h Handler) {
	handlers[method] = h
}

// Start begins the event loop. This function never returns.
func Start() {
	methBuf := make([]byte, 256) // Max method name length

	for {
		// Wait for next command (blocking host call)
		res := next_command(ptr(methBuf), uint32(cap(methBuf)))
		methLen := uint32(res)
		payLen := uint32(res >> 32)

		method := string(methBuf[:methLen])

		// Read payload
		var payload []byte
		if payLen > 0 {
			payload = make([]byte, payLen)
			read_payload(ptr(payload), payLen)
		}

		// Handle
		var result []byte
		var err error

		if h, ok := handlers[method]; ok {
			result, err = h(payload)
		} else {
			err = fmt.Errorf("method %s not found", method)
		}

		if err != nil {
			// Basic error handling: return error string as result for now.
			// Protocol enhancement: separate error channel or flag.
			result = []byte("error: " + err.Error())
		}

		// Submit result
		if len(result) > 0 {
			submit_result(ptr(result), uint32(len(result)))
		} else {
			submit_result(0, 0)
		}
	}
}

// Get retrieves a value from the state store.
func Get(key string) ([]byte, error) {
	keyBytes := []byte(key)

	// Initial buffer size
	bufLen := uint32(1024)
	buf := make([]byte, bufLen)

	actualLen := kv_get(ptr(keyBytes), uint32(len(keyBytes)), ptr(buf), bufLen)

	if actualLen == 0 {
		return nil, nil
	}

	if actualLen > bufLen {
		// Retry with correct size
		bufLen = actualLen
		buf = make([]byte, bufLen)
		actualLen = kv_get(ptr(keyBytes), uint32(len(keyBytes)), ptr(buf), bufLen)
	}

	return buf[:actualLen], nil
}

// Call invokes a method on another service.
func Call(service, method string, payload []byte) ([]byte, error) {
	svcBytes := []byte(service)
	methBytes := []byte(method)

	// Initial buffer size for output
	outLen := uint32(4096)
	out := make([]byte, outLen)

	actualLen := rpc(
		ptr(svcBytes), uint32(len(svcBytes)),
		ptr(methBytes), uint32(len(methBytes)),
		ptr(payload), uint32(len(payload)),
		ptr(out), outLen,
	)

	if actualLen > outLen {
		outLen = actualLen
		out = make([]byte, outLen)
		actualLen = rpc(
			ptr(svcBytes), uint32(len(svcBytes)),
			ptr(methBytes), uint32(len(methBytes)),
			ptr(payload), uint32(len(payload)),
			ptr(out), outLen,
		)
	}

	return out[:actualLen], nil
}

func ptr(b []byte) uint32 {
	if len(b) == 0 {
		return 0
	}
	return uint32(uintptr(unsafe.Pointer(&b[0])))
}
