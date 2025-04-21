package repository

import (
	"context"
	"fmt"
	"time"

	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/internal/domain/repository"
	"api-gateway-sample/pkg/logger"

	"gorm.io/gorm"
)

// ServiceModel represents the service database model
type ServiceModel struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex"`
	Version     string
	Description string
	BaseURL     string
	Timeout     int
	RetryCount  int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EndpointModel represents the endpoint database model
type EndpointModel struct {
	ID           uint `gorm:"primaryKey"`
	ServiceID    string
	Path         string
	Methods      string // Comma-separated list of HTTP methods
	RateLimit    int
	AuthRequired bool
	Timeout      int
	CacheTTL     int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ServiceRepositoryImpl implements the repository.ServiceRepository interface
type ServiceRepositoryImpl struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewServiceRepositoryImpl creates a new ServiceRepositoryImpl instance
func NewServiceRepositoryImpl(db *gorm.DB, logger logger.Logger) repository.ServiceRepository {
	return &ServiceRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// Get retrieves a service by ID
func (r *ServiceRepositoryImpl) Get(ctx context.Context, id string) (*entity.Service, error) {
	var model ServiceModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	service := r.mapModelToEntity(&model)
	if err := r.loadEndpoints(ctx, service); err != nil {
		return nil, err
	}

	return service, nil
}

// GetByID retrieves a service by ID (alias for Get)
func (r *ServiceRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.Service, error) {
	return r.Get(ctx, id)
}

// GetAll retrieves all services
func (r *ServiceRepositoryImpl) GetAll(ctx context.Context) ([]*entity.Service, error) {
	var models []ServiceModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	services := make([]*entity.Service, len(models))
	for i, model := range models {
		service := r.mapModelToEntity(&model)
		if err := r.loadEndpoints(ctx, service); err != nil {
			return nil, err
		}
		services[i] = service
	}

	return services, nil
}

// Create creates a new service
func (r *ServiceRepositoryImpl) Create(ctx context.Context, service *entity.Service) error {
	model := r.mapEntityToModel(service)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	for _, endpoint := range service.Endpoints {
		endpointModel := r.mapEndpointToModel(&endpoint, service.ID)
		if err := r.db.WithContext(ctx).Create(&endpointModel).Error; err != nil {
			return fmt.Errorf("failed to create endpoint: %w", err)
		}
	}

	return nil
}

// Update updates an existing service
func (r *ServiceRepositoryImpl) Update(ctx context.Context, service *entity.Service) error {
	model := r.mapEntityToModel(service)
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	// Delete existing endpoints
	if err := r.db.WithContext(ctx).Where("service_id = ?", service.ID).Delete(&EndpointModel{}).Error; err != nil {
		return fmt.Errorf("failed to delete endpoints: %w", err)
	}

	// Create new endpoints
	for _, endpoint := range service.Endpoints {
		endpointModel := r.mapEndpointToModel(&endpoint, service.ID)
		if err := r.db.WithContext(ctx).Create(&endpointModel).Error; err != nil {
			return fmt.Errorf("failed to create endpoint: %w", err)
		}
	}

	return nil
}

// Delete deletes a service by ID
func (r *ServiceRepositoryImpl) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("service_id = ?", id).Delete(&EndpointModel{}).Error; err != nil {
		return fmt.Errorf("failed to delete endpoints: %w", err)
	}

	if err := r.db.WithContext(ctx).Delete(&ServiceModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	return nil
}

// FindByName finds a service by name
func (r *ServiceRepositoryImpl) FindByName(ctx context.Context, name string) (*entity.Service, error) {
	var model ServiceModel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to find service: %w", err)
	}

	service := r.mapModelToEntity(&model)
	if err := r.loadEndpoints(ctx, service); err != nil {
		return nil, err
	}

	return service, nil
}

// GetByEndpoint finds services by endpoint path and method
func (r *ServiceRepositoryImpl) GetByEndpoint(ctx context.Context, path string, method string) ([]*entity.Service, error) {
	var models []ServiceModel
	if err := r.db.WithContext(ctx).
		Joins("JOIN endpoints ON endpoints.service_id = services.id").
		Where("endpoints.path = ? AND endpoints.methods LIKE ?", path, "%"+method+"%").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	services := make([]*entity.Service, len(models))
	for i, model := range models {
		service := r.mapModelToEntity(&model)
		if err := r.loadEndpoints(ctx, service); err != nil {
			return nil, err
		}
		services[i] = service
	}

	return services, nil
}

// Helper functions

func (r *ServiceRepositoryImpl) mapModelToEntity(model *ServiceModel) *entity.Service {
	return &entity.Service{
		ID:          model.ID,
		Name:        model.Name,
		Version:     model.Version,
		Description: model.Description,
		BaseURL:     model.BaseURL,
		Timeout:     model.Timeout,
		RetryCount:  model.RetryCount,
		IsActive:    model.IsActive,
		Endpoints:   make([]entity.Endpoint, 0),
		Metadata:    make(map[string]string),
	}
}

func (r *ServiceRepositoryImpl) mapEntityToModel(service *entity.Service) *ServiceModel {
	return &ServiceModel{
		ID:          service.ID,
		Name:        service.Name,
		Version:     service.Version,
		Description: service.Description,
		BaseURL:     service.BaseURL,
		Timeout:     service.Timeout,
		RetryCount:  service.RetryCount,
		IsActive:    service.IsActive,
	}
}

func (r *ServiceRepositoryImpl) mapEndpointToModel(endpoint *entity.Endpoint, serviceID string) *EndpointModel {
	return &EndpointModel{
		ServiceID:    serviceID,
		Path:         endpoint.Path,
		Methods:      fmt.Sprintf("%v", endpoint.Methods), // Convert slice to string
		RateLimit:    endpoint.RateLimit,
		AuthRequired: endpoint.AuthRequired,
		Timeout:      endpoint.Timeout,
	}
}

func (r *ServiceRepositoryImpl) loadEndpoints(ctx context.Context, service *entity.Service) error {
	var models []EndpointModel
	if err := r.db.WithContext(ctx).Where("service_id = ?", service.ID).Find(&models).Error; err != nil {
		return fmt.Errorf("failed to load endpoints: %w", err)
	}

	for _, model := range models {
		endpoint := entity.Endpoint{
			Path:         model.Path,
			Methods:      []string{}, // Parse methods string to slice
			RateLimit:    model.RateLimit,
			AuthRequired: model.AuthRequired,
			Timeout:      model.Timeout,
		}
		service.AddEndpoint(endpoint)
	}

	return nil
}
