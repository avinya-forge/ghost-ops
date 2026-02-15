package main

import (
	"fmt"
	"ghost-ops/pkg/sdk/guest"
)

func Handle(payload []byte) ([]byte, error) {
	key := string(payload)

	val, err := guest.Get(key)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return []byte("not found"), nil
	}

	result := fmt.Sprintf("value: %s", string(val))
	return []byte(result), nil
}

func main() {
	guest.Register("Handle", Handle)
	guest.Start()
}
