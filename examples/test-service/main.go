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
