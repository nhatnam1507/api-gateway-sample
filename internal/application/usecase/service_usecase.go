package usecase

import (
	"context"

	"api-gateway-sample/internal/application/dto"
	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/internal/domain/repository"
	"api-gateway-sample/pkg/errors"
)

// ServiceUseCase handles service-related business logic
type ServiceUseCase struct {
	serviceRepo repository.ServiceRepository
	cache       repository.CacheRepository
}

// NewServiceUseCase creates a new ServiceUseCase instance
func NewServiceUseCase(serviceRepo repository.ServiceRepository, cache repository.CacheRepository) *ServiceUseCase {
	return &ServiceUseCase{
		serviceRepo: serviceRepo,
		cache:       cache,
	}
}

// CreateService creates a new service
func (uc *ServiceUseCase) CreateService(ctx context.Context, req *dto.CreateServiceRequest) (*dto.ServiceResponse, error) {
	// Check if service with the same name already exists
	if _, err := uc.serviceRepo.FindByName(ctx, req.Name); err == nil {
		return nil, errors.ErrAlreadyExists
	} else if !errors.IsNotFound(err) {
		return nil, err
	}

	// Convert request to entity
	service := req.ToEntity()

	// Create service
	if err := uc.serviceRepo.Create(ctx, service); err != nil {
		return nil, err
	}

	// Convert entity to response
	return dto.FromEntity(service), nil
}

// GetService retrieves a service by ID
func (uc *ServiceUseCase) GetService(ctx context.Context, id string) (*dto.ServiceResponse, error) {
	service, err := uc.serviceRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return dto.FromEntity(service), nil
}

// UpdateService updates an existing service
func (uc *ServiceUseCase) UpdateService(ctx context.Context, id string, req *dto.UpdateServiceRequest) (*dto.ServiceResponse, error) {
	// Check if service exists
	service, err := uc.serviceRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if new name is already taken by another service
	if req.Name != service.Name {
		if existing, err := uc.serviceRepo.FindByName(ctx, req.Name); err == nil && existing.ID != id {
			return nil, errors.ErrAlreadyExists
		} else if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
	}

	// Update service fields
	service.Name = req.Name
	service.BaseURL = req.BaseURL
	service.Endpoints = make([]entity.Endpoint, len(req.Endpoints))
	for i, e := range req.Endpoints {
		service.Endpoints[i] = entity.Endpoint{
			Path:         e.Path,
			Methods:      e.Methods,
			RateLimit:    e.RateLimit,
			AuthRequired: e.AuthRequired,
			Timeout:      e.Timeout,
			RetryCount:   e.RetryCount,
			RetryDelay:   e.RetryDelay,
			CircuitBreaker: struct {
				Enabled          bool    `json:"enabled"`
				FailureThreshold float64 `json:"failureThreshold"`
				MinRequestCount  int     `json:"minRequestCount"`
				BreakDuration    int     `json:"breakDuration"`
				HalfOpenRequests int     `json:"halfOpenRequests"`
			}{
				Enabled:          e.CircuitBreaker.Enabled,
				FailureThreshold: e.CircuitBreaker.FailureThreshold,
				MinRequestCount:  e.CircuitBreaker.MinRequestCount,
				BreakDuration:    e.CircuitBreaker.BreakDuration,
				HalfOpenRequests: e.CircuitBreaker.HalfOpenRequests,
			},
			Cache: struct {
				Enabled bool `json:"enabled"`
				TTL     int  `json:"ttl"`
			}{
				Enabled: e.Cache.Enabled,
				TTL:     e.Cache.TTL,
			},
			Transform: struct {
				Request  map[string]string `json:"request"`
				Response map[string]string `json:"response"`
			}{
				Request:  e.Transform.Request,
				Response: e.Transform.Response,
			},
		}
	}

	// Update service
	if err := uc.serviceRepo.Update(ctx, service); err != nil {
		return nil, err
	}

	return dto.FromEntity(service), nil
}

// DeleteService deletes a service by ID
func (uc *ServiceUseCase) DeleteService(ctx context.Context, id string) error {
	return uc.serviceRepo.Delete(ctx, id)
}

// ListServices retrieves all services
func (uc *ServiceUseCase) ListServices(ctx context.Context) ([]*dto.ServiceResponse, error) {
	services, err := uc.serviceRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.ServiceResponse, len(services))
	for i, service := range services {
		responses[i] = dto.FromEntity(service)
	}

	return responses, nil
}

// FindServiceByName finds a service by name
func (uc *ServiceUseCase) FindServiceByName(ctx context.Context, name string) (*dto.ServiceResponse, error) {
	service, err := uc.serviceRepo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return dto.FromEntity(service), nil
}
