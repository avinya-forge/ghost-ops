package resilience

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter.
type RateLimiter struct {
	mu           sync.Mutex
	rate         float64 // tokens per second
	burst        int
	tokens       float64
	lastUpdate   time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

// Allow checks if a request is allowed.
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()
	rl.lastUpdate = now

	// Add tokens
	rl.tokens += elapsed * rl.rate
	if rl.tokens > float64(rl.burst) {
		rl.tokens = float64(rl.burst)
	}

	// Consume token
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}
