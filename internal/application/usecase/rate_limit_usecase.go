package usecase

import (
	"context"

	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/internal/domain/service"
	"api-gateway-sample/pkg/logger"
)

// RateLimitUseCase implements the use case for rate limiting
type RateLimitUseCase struct {
	rateLimitService service.RateLimitService
	logger           logger.Logger
}

// NewRateLimitUseCase creates a new RateLimitUseCase instance
func NewRateLimitUseCase(rateLimitService service.RateLimitService, logger logger.Logger) *RateLimitUseCase {
	return &RateLimitUseCase{
		rateLimitService: rateLimitService,
		logger:           logger,
	}
}

// CheckLimit checks if a request exceeds the rate limit
func (uc *RateLimitUseCase) CheckLimit(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) (bool, error) {
	return uc.rateLimitService.CheckLimit(ctx, request, service, endpoint)
}

// GetLimit gets the current rate limit for a client
func (uc *RateLimitUseCase) GetLimit(ctx context.Context, clientID string, service *entity.Service, endpoint *entity.Endpoint) (int, int, error) {
	return uc.rateLimitService.GetLimit(ctx, clientID, service, endpoint)
}
