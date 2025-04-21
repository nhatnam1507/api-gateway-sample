package usecase

import (
	"context"
	"fmt"

	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/internal/domain/repository"
	"api-gateway-sample/internal/domain/service"
	"api-gateway-sample/pkg/errors"
	"api-gateway-sample/pkg/logger"
)

// ProxyUseCase implements the use case for proxying requests
type ProxyUseCase struct {
	serviceRepo      repository.ServiceRepository
	gatewayService   service.GatewayService
	authService      service.AuthService
	rateLimitService service.RateLimitService
	cacheService     service.CacheService
	logger           logger.Logger
}

// NewProxyUseCase creates a new ProxyUseCase instance
func NewProxyUseCase(
	serviceRepo repository.ServiceRepository,
	gatewayService service.GatewayService,
	authService service.AuthService,
	rateLimitService service.RateLimitService,
	cacheService service.CacheService,
	logger logger.Logger,
) *ProxyUseCase {
	return &ProxyUseCase{
		serviceRepo:      serviceRepo,
		gatewayService:   gatewayService,
		authService:      authService,
		rateLimitService: rateLimitService,
		cacheService:     cacheService,
		logger:           logger,
	}
}

// ProxyRequest proxies a request to a backend service
func (uc *ProxyUseCase) ProxyRequest(ctx context.Context, request *entity.Request) (*entity.Response, error) {
	// Validate request
	if err := uc.gatewayService.ValidateRequest(ctx, request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Find service by endpoint path and method
	services, err := uc.serviceRepo.GetByEndpoint(ctx, request.Path, request.Method)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, errors.ErrServiceNotFound
	}

	// For now, we'll use the first matching service
	service := services[0]

	// Find matching endpoint
	var endpoint *entity.Endpoint
	for _, e := range service.Endpoints {
		if e.Path == request.Path {
			endpoint = &e
			break
		}
	}

	if endpoint == nil {
		return nil, fmt.Errorf("no endpoint found for path: %s", request.Path)
	}

	// Check authentication if required
	if endpoint.AuthRequired {
		authenticated, userID, err := uc.authService.Authenticate(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}

		if !authenticated {
			return nil, fmt.Errorf("unauthorized")
		}

		// Set authenticated user ID
		request.UserID = userID

		// Authorize the request
		if err := uc.authService.Authorize(ctx, request, service, endpoint); err != nil {
			return nil, fmt.Errorf("authorization failed: %w", err)
		}
	}

	// Check rate limit
	if endpoint.RateLimit > 0 {
		allowed, err := uc.rateLimitService.CheckLimit(ctx, request, service, endpoint)
		if err != nil {
			return nil, fmt.Errorf("rate limit check failed: %w", err)
		}

		if !allowed {
			return nil, fmt.Errorf("rate limit exceeded")
		}

		// Record the request for rate limiting
		if err := uc.rateLimitService.RecordRequest(ctx, request, service, endpoint); err != nil {
			uc.logger.Warn("Failed to record request for rate limiting", "error", err)
		}
	}

	// Check cache
	if endpoint.CacheTTL > 0 {
		cacheKey := fmt.Sprintf("%s:%s:%s", service.ID, request.Path, request.Method)
		value, found, err := uc.cacheService.Get(ctx, cacheKey)
		if err == nil && found {
			if response, ok := value.(*entity.Response); ok {
				response.CachedResult = true
				return response, nil
			}
		}
	}

	// Transform request
	transformedRequest, err := uc.gatewayService.TransformRequest(ctx, request, service)
	if err != nil {
		return nil, fmt.Errorf("failed to transform request: %w", err)
	}

	// Route request to backend service
	response, err := uc.gatewayService.RouteRequest(ctx, transformedRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to route request: %w", err)
	}

	// Transform response
	transformedResponse, err := uc.gatewayService.TransformResponse(ctx, response, service)
	if err != nil {
		return nil, fmt.Errorf("failed to transform response: %w", err)
	}

	// Cache response if needed
	if endpoint.CacheTTL > 0 {
		cacheKey := fmt.Sprintf("%s:%s:%s", service.ID, request.Path, request.Method)
		if err := uc.cacheService.Set(ctx, cacheKey, transformedResponse, 0); err != nil {
			uc.logger.Warn("Failed to cache response", "error", err)
		}
	}

	return transformedResponse, nil
}
