package store

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"ghost-ops/pkg/protocol"
	"github.com/redis/go-redis/v9"
)

// RedisStore implements protocol.StateStore using Redis.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a new RedisStore.
// It initializes a Redis client and verifies the connection with a Ping.
func NewRedisStore(addr, password string, db int) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Ping to verify connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisStore{
		client: client,
	}, nil
}

// compress gzips the data.
func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decompress un-gzips the data if it has the gzip magic header.
// Otherwise, it returns the data as is.
func decompress(data []byte) ([]byte, error) {
	if len(data) < 2 {
		return data, nil
	}
	// Check for GZIP magic header 0x1f 0x8b
	if data[0] == 0x1f && data[1] == 0x8b {
		r, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer r.Close()
		return io.ReadAll(r)
	}
	return data, nil
}

// GetService retrieves a service record by ID.
func (s *RedisStore) GetService(ctx context.Context, serviceID string) (*protocol.ServiceRecord, error) {
	key := fmt.Sprintf("service:%s", serviceID)
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, err
	}

	rawBytes := []byte(val)
	decompressed, err := decompress(rawBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress service record: %w", err)
	}

	var record protocol.ServiceRecord
	if err := json.Unmarshal(decompressed, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service record: %w", err)
	}
	return &record, nil
}

// UpdateService updates or creates a service record.
func (s *RedisStore) UpdateService(ctx context.Context, record protocol.ServiceRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal service record: %w", err)
	}

	compressed, err := compress(data)
	if err != nil {
		return fmt.Errorf("failed to compress service record: %w", err)
	}

	key := fmt.Sprintf("service:%s", record.ServiceID)
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, key, compressed, 0)
	pipe.SAdd(ctx, "services", record.ServiceID)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// ListServices returns all services.
func (s *RedisStore) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	ids, err := s.client.SMembers(ctx, "services").Result()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []protocol.ServiceRecord{}, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf("service:%s", id)
	}

	vals, err := s.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	var records []protocol.ServiceRecord
	for i, val := range vals {
		if val == nil {
			slog.Warn("ListServices: nil value from Redis MGet, key may have expired", "key", keys[i])
			continue
		}
		strVal, ok := val.(string)
		if !ok {
			slog.Warn("ListServices: unexpected value type from Redis MGet, skipping record",
				"key", keys[i], "type", fmt.Sprintf("%T", val))
			continue
		}

		rawBytes := []byte(strVal)
		decompressed, err := decompress(rawBytes)
		if err != nil {
			slog.Warn("ListServices: failed to decompress record, skipping", "key", keys[i], "error", err)
			continue
		}

		var record protocol.ServiceRecord
		if err := json.Unmarshal(decompressed, &record); err != nil {
			slog.Warn("ListServices: failed to unmarshal record, skipping", "key", keys[i], "error", err)
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

// Get retrieves a value by key.
func (s *RedisStore) Get(ctx context.Context, key string) ([]byte, error) {
	k := fmt.Sprintf("kv:%s", key)
	val, err := s.client.Get(ctx, k).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

// Set stores a value by key.
func (s *RedisStore) Set(ctx context.Context, key string, value []byte) error {
	k := fmt.Sprintf("kv:%s", key)
	return s.client.Set(ctx, k, value, 0).Err()
}

// AcquireLock acquires a distributed lock.
// It returns true if the lock was acquired, false otherwise.
func (s *RedisStore) AcquireLock(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	k := fmt.Sprintf("lock:%s", key)
	err := s.client.SetArgs(ctx, k, value, redis.SetArgs{
		TTL:  ttl,
		Mode: "NX",
	}).Err()
	if err == redis.Nil {
		return false, nil // Not acquired
	} else if err != nil {
		return false, err
	}
	return true, nil // Acquired
}

// ReleaseLock releases a distributed lock if the value matches.
func (s *RedisStore) ReleaseLock(ctx context.Context, key string, value string) error {
	k := fmt.Sprintf("lock:%s", key)
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	_, err := s.client.Eval(ctx, script, []string{k}, value).Result()
	return err
}

// Publish publishes an event to the event bus.
func (s *RedisStore) Publish(ctx context.Context, event protocol.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	// Channel name pattern: event:<EventType>
	channel := fmt.Sprintf("event:%s", event.Type)
	return s.client.Publish(ctx, channel, data).Err()
}

// Subscribe subscribes to events of a specific type.
func (s *RedisStore) Subscribe(ctx context.Context, eventType protocol.EventType) (<-chan protocol.Event, error) {
	channel := fmt.Sprintf("event:%s", eventType)
	pubsub := s.client.Subscribe(ctx, channel)

	// Verify connection
	if _, err := pubsub.Receive(ctx); err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	ch := make(chan protocol.Event)

	// Start a goroutine to read messages
	go func() {
		defer close(ch)
		defer pubsub.Close()

		chRes := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-chRes:
				if !ok {
					return
				}
				var event protocol.Event
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					// Log error? For now just skip
					continue
				}
				select {
				case ch <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}
