package store

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ghost-ops/pkg/protocol"
)

func TestRedisStore(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	store, err := NewRedisStore(mr.Addr(), "", 0)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Service CRUD", func(t *testing.T) {
		record := protocol.ServiceRecord{
			ServiceID: "test-service",
			Version:   1,
			WASMHash:  "abc",
			CurrentState: protocol.StateActive,
			SynthesisTimestamp: time.Now().UTC(),
		}

		// Update (Create)
		err := store.UpdateService(ctx, record)
		require.NoError(t, err)

		// Get
		got, err := store.GetService(ctx, "test-service")
		require.NoError(t, err)
		assert.Equal(t, record.ServiceID, got.ServiceID)
		assert.Equal(t, record.WASMHash, got.WASMHash)

		// List
		list, err := store.ListServices(ctx)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, record.ServiceID, list[0].ServiceID)

		// Update existing
		record.Version = 2
		err = store.UpdateService(ctx, record)
		require.NoError(t, err)

		got, err = store.GetService(ctx, "test-service")
		require.NoError(t, err)
		assert.Equal(t, 2, got.Version)
	})

	t.Run("Service Not Found", func(t *testing.T) {
		got, err := store.GetService(ctx, "non-existent")
		require.NoError(t, err) // Should be nil error
		assert.Nil(t, got)      // Should be nil record
	})

	t.Run("KV CRUD", func(t *testing.T) {
		key := "config:key"
		val := []byte("value")

		// Set
		err := store.Set(ctx, key, val)
		require.NoError(t, err)

		// Get
		got, err := store.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, val, got)

		// Get Not Found
		_, err = store.Get(ctx, "unknown")
		assert.Error(t, err)
	})

	t.Run("Distributed Lock", func(t *testing.T) {
		lockKey := "resource-1"
		lockVal := "client-uuid"
		ttl := 1 * time.Second

		// Acquire
		acquired, err := store.AcquireLock(ctx, lockKey, lockVal, ttl)
		require.NoError(t, err)
		assert.True(t, acquired)

		// Acquire again (should fail)
		acquired, err = store.AcquireLock(ctx, lockKey, "other-client", ttl)
		require.NoError(t, err)
		assert.False(t, acquired)

		// Release with wrong value (should not release)
		err = store.ReleaseLock(ctx, lockKey, "wrong-client")
		require.NoError(t, err)

		// Verify still locked
		acquired, err = store.AcquireLock(ctx, lockKey, "other-client", ttl)
		require.NoError(t, err) // Fixed: Added require.NoError
		assert.False(t, acquired)

		// Release with correct value
		err = store.ReleaseLock(ctx, lockKey, lockVal)
		require.NoError(t, err)

		// Acquire again (should succeed)
		acquired, err = store.AcquireLock(ctx, lockKey, "other-client", ttl)
		require.NoError(t, err)
		assert.True(t, acquired)

		// FastForward to expire
		mr.FastForward(ttl + 100*time.Millisecond)

		// Acquire after expiry (should succeed)
		acquired, err = store.AcquireLock(ctx, lockKey, "new-client", ttl)
		require.NoError(t, err)
		assert.True(t, acquired)
	})
}
