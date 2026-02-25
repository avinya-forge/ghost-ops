package resilience

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cb := NewCircuitBreaker(2, 2, 100*time.Millisecond)

	// Closed -> Open
	assert.Equal(t, StateClosed, cb.State())

	// Fail 1
	cb.Execute(func() error { return errors.New("fail") })
	assert.Equal(t, StateClosed, cb.State())

	// Fail 2
	cb.Execute(func() error { return errors.New("fail") })
	assert.Equal(t, StateOpen, cb.State())

	// Open -> Open (blocked)
	err := cb.Execute(func() error { return nil })
	assert.Equal(t, ErrCircuitOpen, err)

	// Wait for reset -> Half-Open
	time.Sleep(150 * time.Millisecond)

	// First success in Half-Open
	cb.Execute(func() error { return nil })
	assert.Equal(t, StateHalfOpen, cb.State())

	// Second success -> Closed
	cb.Execute(func() error { return nil })
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_HalfOpen_Fail(t *testing.T) {
	cb := NewCircuitBreaker(1, 1, 100*time.Millisecond)

	cb.Execute(func() error { return errors.New("fail") })
	assert.Equal(t, StateOpen, cb.State())

	time.Sleep(150 * time.Millisecond)

	// Half-Open -> Fail -> Open
	cb.Execute(func() error { return errors.New("fail") })
	assert.Equal(t, StateOpen, cb.State())
}
