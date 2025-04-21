package repository

import (
	"context"
	"time"
)

// CacheRepository defines the interface for cache operations
type CacheRepository interface {
	// Set stores a value in the cache with the specified TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Get retrieves a value from the cache
	Get(ctx context.Context, key string, value interface{}) error

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error

	// SetNX sets a value in the cache only if the key does not exist
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)

	// GetWithTTL retrieves a value and its remaining TTL from the cache
	GetWithTTL(ctx context.Context, key string, value interface{}) (time.Duration, error)

	// UpdateTTL updates the TTL of an existing key
	UpdateTTL(ctx context.Context, key string, ttl time.Duration) error

	// Clear removes all keys matching the pattern
	Clear(ctx context.Context, pattern string) error

	// Ping checks the connection to the cache
	Ping(ctx context.Context) error

	// Close closes the cache connection
	Close() error
}
