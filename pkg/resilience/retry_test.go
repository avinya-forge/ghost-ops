package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry_Success(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	err := Retry(ctx, func() error {
		attempts++
		if attempts < 2 {
			return errors.New("fail")
		}
		return nil
	}, WithMaxAttempts(3), WithBackoff(1*time.Millisecond, 10*time.Millisecond))

	assert.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestRetry_Failure(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	err := Retry(ctx, func() error {
		attempts++
		return errors.New("fail")
	}, WithMaxAttempts(3), WithBackoff(1*time.Millisecond, 10*time.Millisecond))

	assert.Error(t, err)
	assert.Equal(t, 3, attempts)
	assert.Contains(t, err.Error(), "retry failed after 3 attempts")
}

func TestRetry_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := Retry(ctx, func() error {
		attempts++
		return errors.New("fail")
	}, WithMaxAttempts(5), WithBackoff(50*time.Millisecond, 100*time.Millisecond))

	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}
