package cache

import (
	"context"
	"testing"
	"time"

	"api-gateway-sample/internal/domain/repository"
	"api-gateway-sample/pkg/errors"

	"github.com/stretchr/testify/assert"
)

// MockCache is a simple in-memory cache for testing
type MockCache struct {
	data map[string]interface{}
}

// Ensure MockCache implements CacheRepository
var _ repository.CacheRepository = (*MockCache)(nil)

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]interface{}),
	}
}

func (c *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.data[key] = value
	return nil
}

func (c *MockCache) Get(ctx context.Context, key string, value interface{}) error {
	if val, ok := c.data[key]; ok {
		// For testing purposes, just assign the value directly
		*(value.(*interface{})) = val
		return nil
	}
	return errors.ErrNotFound
}

func (c *MockCache) Delete(ctx context.Context, key string) error {
	delete(c.data, key)
	return nil
}

func (c *MockCache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	if _, ok := c.data[key]; ok {
		return false, nil
	}
	c.data[key] = value
	return true, nil
}

func (c *MockCache) GetWithTTL(ctx context.Context, key string, value interface{}) (time.Duration, error) {
	if val, ok := c.data[key]; ok {
		*(value.(*interface{})) = val
		return 1 * time.Hour, nil
	}
	return 0, errors.ErrNotFound
}

func (c *MockCache) UpdateTTL(ctx context.Context, key string, ttl time.Duration) error {
	if _, ok := c.data[key]; ok {
		return nil
	}
	return errors.ErrNotFound
}

func (c *MockCache) Clear(ctx context.Context, pattern string) error {
	// For simplicity, clear all data
	c.data = make(map[string]interface{})
	return nil
}

func (c *MockCache) Ping(ctx context.Context) error {
	return nil
}

func (c *MockCache) Close() error {
	return nil
}

func TestCacheServiceWithMock(t *testing.T) {
	// Create a mock cache
	mockCache := NewMockCache()

	// Create a cache service adapter
	cacheService := &CacheServiceAdapter{
		cache: mockCache,
	}

	// Test Set
	t.Run("Set", func(t *testing.T) {
		err := cacheService.Set(context.Background(), "test-key", "test-value", 1*time.Hour)
		assert.NoError(t, err)
	})

	// Test Get
	t.Run("Get", func(t *testing.T) {
		// Set a value first
		err := cacheService.Set(context.Background(), "test-key", "test-value", 1*time.Hour)
		assert.NoError(t, err)

		// Get the value
		value, found, err := cacheService.Get(context.Background(), "test-key")
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "test-value", value)
	})

	// Test Get Not Found
	t.Run("GetNotFound", func(t *testing.T) {
		value, found, err := cacheService.Get(context.Background(), "non-existent-key")
		assert.Error(t, err)
		assert.False(t, found)
		assert.Nil(t, value)
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		// Set a value first
		err := cacheService.Set(context.Background(), "test-key", "test-value", 1*time.Hour)
		assert.NoError(t, err)

		// Delete the value
		err = cacheService.Delete(context.Background(), "test-key")
		assert.NoError(t, err)

		// Verify it's deleted
		_, found, err := cacheService.Get(context.Background(), "test-key")
		assert.Error(t, err)
		assert.False(t, found)
	})

	// Test Clear
	t.Run("Clear", func(t *testing.T) {
		// Set some values
		err := cacheService.Set(context.Background(), "key1", "value1", 1*time.Hour)
		assert.NoError(t, err)
		err = cacheService.Set(context.Background(), "key2", "value2", 1*time.Hour)
		assert.NoError(t, err)

		// Clear the cache
		err = cacheService.Clear(context.Background())
		assert.NoError(t, err)

		// Verify all values are cleared
		_, found, err := cacheService.Get(context.Background(), "key1")
		assert.Error(t, err)
		assert.False(t, found)

		_, found, err = cacheService.Get(context.Background(), "key2")
		assert.Error(t, err)
		assert.False(t, found)
	})
}
