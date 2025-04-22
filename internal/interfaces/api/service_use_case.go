package api

import (
	"context"

	"api-gateway-sample/internal/application/dto"
)

// ServiceUseCase defines the interface for service use cases
type ServiceUseCase interface {
	CreateService(ctx context.Context, req *dto.CreateServiceRequest) (*dto.ServiceResponse, error)
	GetService(ctx context.Context, id string) (*dto.ServiceResponse, error)
	UpdateService(ctx context.Context, id string, req *dto.UpdateServiceRequest) (*dto.ServiceResponse, error)
	DeleteService(ctx context.Context, id string) error
	ListServices(ctx context.Context) ([]*dto.ServiceResponse, error)
	FindServiceByName(ctx context.Context, name string) (*dto.ServiceResponse, error)
}
