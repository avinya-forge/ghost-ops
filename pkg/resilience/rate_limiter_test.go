package resilience

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_Allow(t *testing.T) {
	// Rate 10/s, Burst 1
	rl := NewRateLimiter(10.0, 1)

	// First call should be allowed (burst)
	assert.True(t, rl.Allow())

	// Immediate second call should be denied (burst consumed)
	assert.False(t, rl.Allow())

	// Wait 100ms (1 token)
	time.Sleep(110 * time.Millisecond)
	assert.True(t, rl.Allow())
}

func TestRateLimiter_Burst(t *testing.T) {
	// Rate 1/s, Burst 3
	rl := NewRateLimiter(1.0, 3)

	assert.True(t, rl.Allow())
	assert.True(t, rl.Allow())
	assert.True(t, rl.Allow())
	assert.False(t, rl.Allow())
}
