package resilience

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// RetryOption configures the retry behavior.
type RetryOption func(*retryConfig)

type retryConfig struct {
	maxAttempts int
	initialBackoff time.Duration
	maxBackoff     time.Duration
	jitter         float64
}

// WithMaxAttempts sets the maximum number of attempts.
func WithMaxAttempts(attempts int) RetryOption {
	return func(c *retryConfig) {
		c.maxAttempts = attempts
	}
}

// WithBackoff sets the initial and maximum backoff duration.
func WithBackoff(initial, max time.Duration) RetryOption {
	return func(c *retryConfig) {
		c.initialBackoff = initial
		c.maxBackoff = max
	}
}

// WithJitter sets the jitter factor (0.0 - 1.0).
func WithJitter(jitter float64) RetryOption {
	return func(c *retryConfig) {
		c.jitter = jitter
	}
}

// Retry executes the given function with retries.
func Retry(ctx context.Context, fn func() error, opts ...RetryOption) error {
	config := retryConfig{
		maxAttempts:    3,
		initialBackoff: 100 * time.Millisecond,
		maxBackoff:     1 * time.Second,
		jitter:         0.1,
	}

	for _, opt := range opts {
		opt(&config)
	}

	var err error
	backoff := config.initialBackoff

	for attempt := 1; attempt <= config.maxAttempts; attempt++ {
		if err = fn(); err == nil {
			return nil
		}

		if attempt == config.maxAttempts {
			break
		}

		// Apply jitter
		jitter := time.Duration(float64(backoff) * config.jitter * (rand.Float64()*2 - 1))
		sleep := backoff + jitter
		if sleep < 0 {
			sleep = 0
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}

		backoff *= 2
		if backoff > config.maxBackoff {
			backoff = config.maxBackoff
		}
	}

	return fmt.Errorf("retry failed after %d attempts: %w", config.maxAttempts, err)
}
