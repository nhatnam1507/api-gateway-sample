package service

import (
	"api-gateway-sample/internal/domain/entity"
	"context"
)

// RateLimitService defines the interface for rate limiting service
type RateLimitService interface {
	// CheckLimit checks if a request exceeds the rate limit
	CheckLimit(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) (bool, error)

	// RecordRequest records a request for rate limiting purposes
	RecordRequest(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) error

	// GetLimit gets the current rate limit for a client
	GetLimit(ctx context.Context, clientID string, service *entity.Service, endpoint *entity.Endpoint) (int, int, error)
}
