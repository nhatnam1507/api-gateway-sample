package client

import (
	"context"

	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/pkg/logger"
)

// GatewayService implements the gateway service interface
type GatewayService struct {
	httpClient *HTTPClient
	logger     logger.Logger
}

// NewGatewayService creates a new GatewayService instance
func NewGatewayService(httpClient *HTTPClient, logger logger.Logger) *GatewayService {
	return &GatewayService{
		httpClient: httpClient,
		logger:     logger,
	}
}

// ValidateRequest validates a request before routing
func (s *GatewayService) ValidateRequest(ctx context.Context, request *entity.Request) error {
	// Basic validation - can be extended based on requirements
	if request == nil {
		return ErrInvalidRequest
	}
	if request.Method == "" {
		return ErrInvalidMethod
	}
	if request.Path == "" {
		return ErrInvalidPath
	}
	return nil
}

// TransformRequest transforms a request before sending to backend
func (s *GatewayService) TransformRequest(ctx context.Context, request *entity.Request, service *entity.Service) (*entity.Request, error) {
	// Create a new request with the same data
	transformed := &entity.Request{
		ID:          request.ID,
		Method:      request.Method,
		Path:        request.Path,
		Headers:     request.Headers,
		QueryParams: request.QueryParams,
		Body:        request.Body,
		ClientIP:    request.ClientIP,
		Timestamp:   request.Timestamp,
		UserID:      request.UserID,
	}

	// Add service-specific headers
	if transformed.Headers == nil {
		transformed.Headers = make(map[string][]string)
	}
	transformed.Headers["X-Service-ID"] = []string{service.ID}
	transformed.Headers["X-Service-Name"] = []string{service.Name}

	return transformed, nil
}

// RouteRequest routes a request to a backend service
func (s *GatewayService) RouteRequest(ctx context.Context, request *entity.Request) (*entity.Response, error) {
	// Create a dummy service for now - in real implementation this would come from service discovery
	service := &entity.Service{
		ID:      "dummy",
		Name:    "dummy",
		BaseURL: "http://localhost",
	}
	return s.httpClient.SendRequest(ctx, request, service)
}

// TransformResponse transforms a response before sending to client
func (s *GatewayService) TransformResponse(ctx context.Context, response *entity.Response, service *entity.Service) (*entity.Response, error) {
	// Create a new response with the same data
	transformed := &entity.Response{
		RequestID:    response.RequestID,
		StatusCode:   response.StatusCode,
		Headers:      response.Headers,
		Body:         response.Body,
		ContentType:  response.ContentType,
		Timestamp:    response.Timestamp,
		LatencyMs:    response.LatencyMs,
		CachedResult: response.CachedResult,
	}

	// Add service-specific headers
	if transformed.Headers == nil {
		transformed.Headers = make(map[string][]string)
	}
	transformed.Headers["X-Service-ID"] = []string{service.ID}
	transformed.Headers["X-Service-Name"] = []string{service.Name}

	return transformed, nil
}

// HandleError handles errors during request processing
func (s *GatewayService) HandleError(ctx context.Context, err error, request *entity.Request) (*entity.Response, error) {
	s.logger.Error("Request processing error",
		"error", err,
		"method", request.Method,
		"path", request.Path,
	)
	return nil, err
}
