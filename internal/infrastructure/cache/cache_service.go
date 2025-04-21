package cache

import (
	"context"
	"time"

	"api-gateway-sample/internal/domain/repository"
	"api-gateway-sample/internal/domain/service"
)

// CacheServiceAdapter adapts the CacheRepository to implement CacheService
type CacheServiceAdapter struct {
	cache repository.CacheRepository
}

// NewCacheService creates a new CacheService instance
func NewCacheService(cache repository.CacheRepository) service.CacheService {
	return &CacheServiceAdapter{
		cache: cache,
	}
}

// Get retrieves a value from the cache
func (s *CacheServiceAdapter) Get(ctx context.Context, key string) (interface{}, bool, error) {
	var value interface{}
	err := s.cache.Get(ctx, key, &value)
	if err != nil {
		return nil, false, err
	}
	return value, true, nil
}

// Set stores a value in the cache with an optional TTL
func (s *CacheServiceAdapter) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.cache.Set(ctx, key, value, ttl)
}

// Delete removes a value from the cache
func (s *CacheServiceAdapter) Delete(ctx context.Context, key string) error {
	return s.cache.Delete(ctx, key)
}

// Clear removes all values from the cache
func (s *CacheServiceAdapter) Clear(ctx context.Context) error {
	return s.cache.Clear(ctx, "*")
}
