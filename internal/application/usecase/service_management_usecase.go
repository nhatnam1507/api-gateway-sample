package usecase

import (
	"context"

	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/internal/domain/repository"
	"api-gateway-sample/pkg/logger"
)

// ServiceManagementUseCase implements the use case for managing services
type ServiceManagementUseCase struct {
	serviceRepo repository.ServiceRepository
	logger      logger.Logger
}

// NewServiceManagementUseCase creates a new ServiceManagementUseCase instance
func NewServiceManagementUseCase(serviceRepo repository.ServiceRepository, logger logger.Logger) *ServiceManagementUseCase {
	return &ServiceManagementUseCase{
		serviceRepo: serviceRepo,
		logger:      logger,
	}
}

// GetAllServices returns all registered services
func (uc *ServiceManagementUseCase) GetAllServices(ctx context.Context) ([]*entity.Service, error) {
	return uc.serviceRepo.GetAll(ctx)
}

// GetServiceByID returns a service by its ID
func (uc *ServiceManagementUseCase) GetServiceByID(ctx context.Context, id string) (*entity.Service, error) {
	return uc.serviceRepo.GetByID(ctx, id)
}

// CreateService creates a new service
func (uc *ServiceManagementUseCase) CreateService(ctx context.Context, service *entity.Service) error {
	return uc.serviceRepo.Create(ctx, service)
}

// UpdateService updates an existing service
func (uc *ServiceManagementUseCase) UpdateService(ctx context.Context, service *entity.Service) error {
	return uc.serviceRepo.Update(ctx, service)
}

// DeleteService deletes a service by its ID
func (uc *ServiceManagementUseCase) DeleteService(ctx context.Context, id string) error {
	return uc.serviceRepo.Delete(ctx, id)
}
