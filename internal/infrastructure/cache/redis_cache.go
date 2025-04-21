package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"api-gateway-sample/internal/domain/repository"
	"api-gateway-sample/pkg/errors"
)

// RedisCache implements the repository.CacheRepository interface
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new RedisCache instance
func NewRedisCache(client *redis.Client) repository.CacheRepository {
	return &RedisCache{
		client: client,
	}
}

// Set stores a value in the cache with the specified TTL
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache value: %w", err)
	}

	return nil
}

// Get retrieves a value from the cache
func (c *RedisCache) Get(ctx context.Context, key string, value interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return errors.ErrNotFound
		}
		return fmt.Errorf("failed to get cache value: %w", err)
	}

	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	return nil
}

// Delete removes a value from the cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete cache value: %w", err)
	}

	return nil
}

// SetNX sets a value in the cache only if the key does not exist
func (c *RedisCache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal cache value: %w", err)
	}

	ok, err := c.client.SetNX(ctx, key, data, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set cache value: %w", err)
	}

	return ok, nil
}

// GetWithTTL retrieves a value and its remaining TTL from the cache
func (c *RedisCache) GetWithTTL(ctx context.Context, key string, value interface{}) (time.Duration, error) {
	// Get the value
	if err := c.Get(ctx, key, value); err != nil {
		return 0, err
	}

	// Get the TTL
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	return ttl, nil
}

// UpdateTTL updates the TTL of an existing key
func (c *RedisCache) UpdateTTL(ctx context.Context, key string, ttl time.Duration) error {
	ok, err := c.client.Expire(ctx, key, ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to update TTL: %w", err)
	}

	if !ok {
		return errors.ErrNotFound
	}

	return nil
}

// Clear removes all keys matching the pattern
func (c *RedisCache) Clear(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", iter.Val(), err)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	return nil
}

// Ping checks the connection to Redis
func (c *RedisCache) Ping(ctx context.Context) error {
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis connection: %w", err)
	}

	return nil
}
