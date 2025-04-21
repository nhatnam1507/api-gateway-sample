package service

import (
	"api-gateway-sample/internal/domain/entity"
	"context"
)

// AuthService defines the interface for authentication service
type AuthService interface {
	// Authenticate authenticates a request
	Authenticate(ctx context.Context, request *entity.Request) (bool, string, error)

	// Authorize authorizes a request for a specific service and endpoint
	Authorize(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) error

	// GenerateToken generates an authentication token
	GenerateToken(ctx context.Context, userID string, claims map[string]interface{}) (string, error)

	// ValidateToken validates an authentication token
	ValidateToken(ctx context.Context, token string) (map[string]interface{}, error)
}
