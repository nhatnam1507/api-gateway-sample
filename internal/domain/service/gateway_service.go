package service

import (
	"api-gateway-sample/internal/domain/entity"
	"context"
)

// GatewayService defines the interface for the API Gateway service
type GatewayService interface {
	// RouteRequest routes a request to the appropriate backend service
	RouteRequest(ctx context.Context, request *entity.Request) (*entity.Response, error)

	// ValidateRequest validates a request before routing
	ValidateRequest(ctx context.Context, request *entity.Request) error

	// TransformRequest transforms a request before sending to backend
	TransformRequest(ctx context.Context, request *entity.Request, service *entity.Service) (*entity.Request, error)

	// TransformResponse transforms a response before sending to client
	TransformResponse(ctx context.Context, response *entity.Response, service *entity.Service) (*entity.Response, error)

	// HandleError handles errors during request processing
	HandleError(ctx context.Context, err error, request *entity.Request) (*entity.Response, error)
}
