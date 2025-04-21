package repository

import (
	"context"

	"api-gateway-sample/internal/domain/entity"
)

// ServiceRepository defines the interface for service operations
type ServiceRepository interface {
	// Create creates a new service
	Create(ctx context.Context, service *entity.Service) error

	// Get retrieves a service by ID
	Get(ctx context.Context, id string) (*entity.Service, error)

	// GetByID retrieves a service by ID (alias for Get)
	GetByID(ctx context.Context, id string) (*entity.Service, error)

	// Update updates an existing service
	Update(ctx context.Context, service *entity.Service) error

	// Delete deletes a service by ID
	Delete(ctx context.Context, id string) error

	// GetAll retrieves all services
	GetAll(ctx context.Context) ([]*entity.Service, error)

	// FindByName finds a service by name
	FindByName(ctx context.Context, name string) (*entity.Service, error)

	// GetByEndpoint finds services by endpoint path and method
	GetByEndpoint(ctx context.Context, path string, method string) ([]*entity.Service, error)
}
