package service

import (
	"context"
	"time"
)

// CacheService defines the interface for caching operations
type CacheService interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) (interface{}, bool, error)

	// Set stores a value in the cache with an optional TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error

	// Clear removes all values from the cache
	Clear(ctx context.Context) error
}
