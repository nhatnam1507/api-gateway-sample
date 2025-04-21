package mock

import (
	"context"
	"sync"

	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/internal/domain/repository"
	"api-gateway-sample/pkg/errors"
)

// ServiceRepositoryMock is a mock implementation of the ServiceRepository interface
type ServiceRepositoryMock struct {
	services map[string]*entity.Service
	mu       sync.RWMutex
}

// NewServiceRepositoryMock creates a new ServiceRepositoryMock instance
func NewServiceRepositoryMock() repository.ServiceRepository {
	return &ServiceRepositoryMock{
		services: make(map[string]*entity.Service),
	}
}

// Create creates a new service
func (r *ServiceRepositoryMock) Create(ctx context.Context, service *entity.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if service with the same name already exists
	for _, s := range r.services {
		if s.Name == service.Name {
			return errors.ErrAlreadyExists
		}
	}

	// Generate ID if not provided
	if service.ID == "" {
		service.ID = "test-id"
	}

	r.services[service.ID] = service
	return nil
}

// Get retrieves a service by ID
func (r *ServiceRepositoryMock) Get(ctx context.Context, id string) (*entity.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, ok := r.services[id]
	if !ok {
		return nil, errors.ErrNotFound
	}

	return service, nil
}

// GetByID retrieves a service by ID (alias for Get)
func (r *ServiceRepositoryMock) GetByID(ctx context.Context, id string) (*entity.Service, error) {
	return r.Get(ctx, id)
}

// Update updates an existing service
func (r *ServiceRepositoryMock) Update(ctx context.Context, service *entity.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.services[service.ID]; !ok {
		return errors.ErrNotFound
	}

	// Check if new name is already taken by another service
	for _, s := range r.services {
		if s.Name == service.Name && s.ID != service.ID {
			return errors.ErrAlreadyExists
		}
	}

	r.services[service.ID] = service
	return nil
}

// Delete deletes a service by ID
func (r *ServiceRepositoryMock) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.services[id]; !ok {
		return errors.ErrNotFound
	}

	delete(r.services, id)
	return nil
}

// GetAll retrieves all services
func (r *ServiceRepositoryMock) GetAll(ctx context.Context) ([]*entity.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*entity.Service, 0, len(r.services))
	for _, service := range r.services {
		services = append(services, service)
	}

	return services, nil
}

// FindByName finds a service by name
func (r *ServiceRepositoryMock) FindByName(ctx context.Context, name string) (*entity.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, service := range r.services {
		if service.Name == name {
			return service, nil
		}
	}

	return nil, errors.ErrNotFound
}

// Reset clears all services (useful for testing)
func (r *ServiceRepositoryMock) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.services = make(map[string]*entity.Service)
}

// GetByEndpoint finds services by endpoint path and method
func (r *ServiceRepositoryMock) GetByEndpoint(ctx context.Context, path string, method string) ([]*entity.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matchingServices []*entity.Service
	for _, service := range r.services {
		for _, endpoint := range service.Endpoints {
			if endpoint.Path == path {
				// Check if the endpoint supports the method
				for _, supportedMethod := range endpoint.Methods {
					if supportedMethod == method || supportedMethod == "*" {
						matchingServices = append(matchingServices, service)
						break
					}
				}
			}
		}
	}

	if len(matchingServices) == 0 {
		return nil, errors.ErrNotFound
	}

	return matchingServices, nil
}
