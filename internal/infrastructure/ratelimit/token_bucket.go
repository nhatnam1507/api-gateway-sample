package ratelimit

import (
	"context"
	"fmt"
	"time"

	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// TokenBucketRateLimiter implements rate limiting using the token bucket algorithm
type TokenBucketRateLimiter struct {
	client *redis.Client
	logger logger.Logger
}

// NewTokenBucketRateLimiter creates a new TokenBucketRateLimiter instance
func NewTokenBucketRateLimiter(client *redis.Client, logger logger.Logger) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		client: client,
		logger: logger,
	}
}

// CheckLimit checks if a request exceeds the rate limit
func (r *TokenBucketRateLimiter) CheckLimit(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s:%s:%s", service.ID, request.Path, request.ClientIP)

	// Get current token count
	count, err := r.client.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, err
	}

	// If key doesn't exist or expired, initialize it
	if err == redis.Nil {
		count = endpoint.RateLimit
	}

	// Check if we have tokens available
	if count <= 0 {
		return false, nil
	}

	return true, nil
}

// RecordRequest records a request for rate limiting purposes
func (r *TokenBucketRateLimiter) RecordRequest(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) error {
	key := fmt.Sprintf("ratelimit:%s:%s:%s", service.ID, request.Path, request.ClientIP)

	// Decrement token count
	count, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		return err
	}

	// Set expiration if this is a new key
	if int(count) == endpoint.RateLimit-1 {
		r.client.Expire(ctx, key, time.Minute)
	}

	return nil
}

// GetLimit gets the current rate limit for a client
func (r *TokenBucketRateLimiter) GetLimit(ctx context.Context, clientID string, service *entity.Service, endpoint *entity.Endpoint) (int, int, error) {
	key := fmt.Sprintf("ratelimit:%s:%s:%s", service.ID, endpoint.Path, clientID)

	// Get current token count
	count, err := r.client.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return 0, 0, err
	}

	if err == redis.Nil {
		count = endpoint.RateLimit
	}

	return count, endpoint.RateLimit, nil
}
