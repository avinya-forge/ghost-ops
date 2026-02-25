package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"ghost-ops/pkg/protocol"
)

// RedisStore implements protocol.StateStore using Redis.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a new RedisStore.
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

	var record protocol.ServiceRecord
	if err := json.Unmarshal([]byte(val), &record); err != nil {
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

	key := fmt.Sprintf("service:%s", record.ServiceID)
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, key, data, 0)
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
	for _, val := range vals {
		if val == nil {
			continue
		}
		strVal, ok := val.(string)
		if !ok {
			continue
		}
		var record protocol.ServiceRecord
		if err := json.Unmarshal([]byte(strVal), &record); err != nil {
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
