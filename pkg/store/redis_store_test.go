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

	t.Run("Pub/Sub", func(t *testing.T) {
		eventType := protocol.EventType("test-event")

		// Subscribe
		ch, err := store.Subscribe(ctx, eventType)
		require.NoError(t, err)

		// Publish
		event := protocol.Event{
			Type:      eventType,
			ServiceID: "service-1",
			Payload:   map[string]interface{}{"foo": "bar"},
			Timestamp: time.Now().Unix(),
		}

		// Give subscription a moment to propagate in miniredis
		time.Sleep(50 * time.Millisecond)

		err = store.Publish(ctx, event)
		require.NoError(t, err)

		// Receive
		select {
		case received := <-ch:
			assert.Equal(t, event.ServiceID, received.ServiceID)
			assert.Equal(t, event.Type, received.Type)
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for event")
		}
	})

	t.Run("Compression", func(t *testing.T) {
		record := protocol.ServiceRecord{
			ServiceID: "compressed-service",
			Version:   1,
			WASMHash:  "compressed-hash",
		}

		// Update (should compress)
		err := store.UpdateService(ctx, record)
		require.NoError(t, err)

		// Verify raw data in Redis is compressed (starts with gzip magic bytes 0x1f 0x8b)
		key := "service:compressed-service"
		val, err := store.client.Get(ctx, key).Bytes()
		require.NoError(t, err)
		assert.True(t, len(val) > 2)
		// Check for gzip magic header
		assert.Equal(t, byte(0x1f), val[0])
		assert.Equal(t, byte(0x8b), val[1])

		// Get (should decompress)
		got, err := store.GetService(ctx, "compressed-service")
		require.NoError(t, err)
		assert.Equal(t, record.ServiceID, got.ServiceID)

		// Backward Compatibility: Write uncompressed data manually
		rawJSON, _ := json.Marshal(record)
		err = store.client.Set(ctx, "service:uncompressed-service", rawJSON, 0).Err()
		require.NoError(t, err)

		// Get uncompressed (should succeed)
		gotUncompressed, err := store.GetService(ctx, "uncompressed-service")
		require.NoError(t, err)
		assert.Equal(t, record.ServiceID, gotUncompressed.ServiceID)
	})
}
