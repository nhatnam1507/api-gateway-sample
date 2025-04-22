package cache

import (
	"context"
	"testing"
	"time"

	"api-gateway-sample/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCacheRepository is a mock implementation of the CacheRepository interface
type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheRepository) Get(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockCacheRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheRepository) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	args := m.Called(ctx, key, value, ttl)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheRepository) GetWithTTL(ctx context.Context, key string, value interface{}) (time.Duration, error) {
	args := m.Called(ctx, key, value)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockCacheRepository) UpdateTTL(ctx context.Context, key string, ttl time.Duration) error {
	args := m.Called(ctx, key, ttl)
	return args.Error(0)
}

func (m *MockCacheRepository) Clear(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCacheRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCacheRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestCacheServiceAdapter_Get(t *testing.T) {
	// Setup
	mockRepo := new(MockCacheRepository)
	service := NewCacheService(mockRepo)

	// Test data
	key := "test-key"
	expectedValue := "test-value"

	// Set up expectations
	mockRepo.On("Get", mock.Anything, key, mock.Anything).Run(func(args mock.Arguments) {
		// Set the value in the output parameter
		valuePtr := args.Get(2).(*interface{})
		*valuePtr = expectedValue
	}).Return(nil)

	// Execute
	value, found, err := service.Get(context.Background(), key)

	// Verify
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, expectedValue, value)
	mockRepo.AssertExpectations(t)
}

func TestCacheServiceAdapter_Get_NotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockCacheRepository)
	service := NewCacheService(mockRepo)

	// Test data
	key := "test-key"

	// Set up expectations
	mockRepo.On("Get", mock.Anything, key, mock.Anything).Return(errors.ErrNotFound)

	// Execute
	value, found, err := service.Get(context.Background(), key)

	// Verify
	assert.Error(t, err)
	assert.False(t, found)
	assert.Nil(t, value)
	mockRepo.AssertExpectations(t)
}

func TestCacheServiceAdapter_Set(t *testing.T) {
	// Setup
	mockRepo := new(MockCacheRepository)
	service := NewCacheService(mockRepo)

	// Test data
	key := "test-key"
	value := "test-value"
	ttl := 1 * time.Hour

	// Set up expectations
	mockRepo.On("Set", mock.Anything, key, value, ttl).Return(nil)

	// Execute
	err := service.Set(context.Background(), key, value, ttl)

	// Verify
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCacheServiceAdapter_Delete(t *testing.T) {
	// Setup
	mockRepo := new(MockCacheRepository)
	service := NewCacheService(mockRepo)

	// Test data
	key := "test-key"

	// Set up expectations
	mockRepo.On("Delete", mock.Anything, key).Return(nil)

	// Execute
	err := service.Delete(context.Background(), key)

	// Verify
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCacheServiceAdapter_Clear(t *testing.T) {
	// Setup
	mockRepo := new(MockCacheRepository)
	service := NewCacheService(mockRepo)

	// Set up expectations
	mockRepo.On("Clear", mock.Anything, "*").Return(nil)

	// Execute
	err := service.Clear(context.Background())

	// Verify
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
