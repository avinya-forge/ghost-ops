package store

import (
	"context"
	"encoding/json"
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
		require.NoError(t, err)
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

	t.Run("Service Compression", func(t *testing.T) {
		record := protocol.ServiceRecord{
			ServiceID:          "compressed-service",
			Version:            1,
			WASMHash:           "xyz",
			CurrentState:       protocol.StateActive,
			SynthesisTimestamp: time.Now().UTC(),
		}

		// Update (Should compress)
		err := store.UpdateService(ctx, record)
		require.NoError(t, err)

		// Verify internal storage is compressed (by peeking into Redis via helper)
		// We use store.Get(kv:) helper but keys are service:..., so we use client directly
		rawVal, err := store.client.Get(ctx, "service:compressed-service").Result()
		require.NoError(t, err)
		// Check for gzip magic header
		require.GreaterOrEqual(t, len(rawVal), 2)
		assert.Equal(t, byte(0x1f), rawVal[0])
		assert.Equal(t, byte(0x8b), rawVal[1])

		// Get (Should decompress)
		got, err := store.GetService(ctx, "compressed-service")
		require.NoError(t, err)
		assert.Equal(t, record.ServiceID, got.ServiceID)
		assert.Equal(t, record.WASMHash, got.WASMHash)
	})

	t.Run("Service Backward Compatibility", func(t *testing.T) {
		// Manually insert uncompressed JSON
		record := protocol.ServiceRecord{
			ServiceID:          "legacy-service",
			Version:            1,
			WASMHash:           "legacy",
			CurrentState:       protocol.StateActive,
			SynthesisTimestamp: time.Now().UTC(),
		}
		data, err := json.Marshal(record)
		require.NoError(t, err)

		// Bypass UpdateService to simulate legacy data
		err = store.client.Set(ctx, "service:legacy-service", data, 0).Err()
		require.NoError(t, err)

		// Get (Should detect not compressed and just unmarshal)
		got, err := store.GetService(ctx, "legacy-service")
		require.NoError(t, err)
		assert.Equal(t, record.ServiceID, got.ServiceID)
		assert.Equal(t, record.WASMHash, got.WASMHash)
	})

	t.Run("Pub/Sub", func(t *testing.T) {
		eventType := protocol.EventType("test_event")

		// Subscribe
		subCtx, subCancel := context.WithCancel(ctx)
		defer subCancel()

		ch, err := store.Subscribe(subCtx, eventType)
		require.NoError(t, err)

		// Publish
		event := protocol.Event{
			Type:      eventType,
			ServiceID: "pubsub-service",
			Payload:   map[string]interface{}{"foo": "bar"},
			Timestamp: time.Now().Unix(),
		}

		// Wait a bit to ensure subscription is active (though Subscribe waits for Receive)
		time.Sleep(50 * time.Millisecond)

		err = store.Publish(ctx, event)
		require.NoError(t, err)

		// Receive
		select {
		case got := <-ch:
			assert.Equal(t, event.ServiceID, got.ServiceID)
			assert.Equal(t, event.Type, got.Type)
			assert.Equal(t, "bar", got.Payload["foo"])
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for event")
		}
	})
}
