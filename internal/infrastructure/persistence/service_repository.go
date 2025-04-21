package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/internal/domain/repository"
	"api-gateway-sample/internal/infrastructure/cache"
	"api-gateway-sample/pkg/errors"
)

// ServiceModel represents the database model for a service
type ServiceModel struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex;not null"`
	BaseURL   string `gorm:"not null"`
	Endpoints string `gorm:"type:jsonb;not null"` // JSON array of endpoints
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// ServiceRepository implements the repository.ServiceRepository interface
type ServiceRepository struct {
	db    *gorm.DB
	cache *cache.RedisCache
}

// NewServiceRepository creates a new ServiceRepository instance
func NewServiceRepository(db *gorm.DB, cache *cache.RedisCache) repository.ServiceRepository {
	return &ServiceRepository{
		db:    db,
		cache: cache,
	}
}

// getCacheKey returns the cache key for a service
func (r *ServiceRepository) getCacheKey(id string) string {
	return fmt.Sprintf("service:%s", id)
}

// Create creates a new service
func (r *ServiceRepository) Create(ctx context.Context, service *entity.Service) error {
	endpoints, err := json.Marshal(service.Endpoints)
	if err != nil {
		return fmt.Errorf("failed to marshal endpoints: %w", err)
	}

	model := &ServiceModel{
		Name:      service.Name,
		BaseURL:   service.BaseURL,
		Endpoints: string(endpoints),
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Update cache
	service.ID = fmt.Sprintf("%d", model.ID)
	if err := r.cache.Set(ctx, r.getCacheKey(service.ID), service, 24*time.Hour); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("failed to cache service: %v\n", err)
	}

	return nil
}

// Get retrieves a service by ID
func (r *ServiceRepository) Get(ctx context.Context, id string) (*entity.Service, error) {
	// Try cache first
	var service entity.Service
	if err := r.cache.Get(ctx, r.getCacheKey(id), &service); err == nil {
		return &service, nil
	}

	// Cache miss, get from database
	var model ServiceModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	var endpoints []entity.Endpoint
	if err := json.Unmarshal([]byte(model.Endpoints), &endpoints); err != nil {
		return nil, fmt.Errorf("failed to unmarshal endpoints: %w", err)
	}

	service = entity.Service{
		ID:        fmt.Sprintf("%d", model.ID),
		Name:      model.Name,
		BaseURL:   model.BaseURL,
		Endpoints: endpoints,
	}

	// Update cache
	if err := r.cache.Set(ctx, r.getCacheKey(id), &service, 24*time.Hour); err != nil {
		fmt.Printf("failed to cache service: %v\n", err)
	}

	return &service, nil
}

// Update updates an existing service
func (r *ServiceRepository) Update(ctx context.Context, service *entity.Service) error {
	endpoints, err := json.Marshal(service.Endpoints)
	if err != nil {
		return fmt.Errorf("failed to marshal endpoints: %w", err)
	}

	model := &ServiceModel{
		Name:      service.Name,
		BaseURL:   service.BaseURL,
		Endpoints: string(endpoints),
	}

	if err := r.db.WithContext(ctx).Model(&ServiceModel{}).Where("id = ?", service.ID).Updates(model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound
		}
		return fmt.Errorf("failed to update service: %w", err)
	}

	// Update cache
	if err := r.cache.Set(ctx, r.getCacheKey(service.ID), service, 24*time.Hour); err != nil {
		fmt.Printf("failed to cache service: %v\n", err)
	}

	return nil
}

// Delete deletes a service by ID
func (r *ServiceRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&ServiceModel{}, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound
		}
		return fmt.Errorf("failed to delete service: %w", err)
	}

	// Remove from cache
	if err := r.cache.Delete(ctx, r.getCacheKey(id)); err != nil {
		fmt.Printf("failed to remove service from cache: %v\n", err)
	}

	return nil
}

// GetAll retrieves all services
func (r *ServiceRepository) GetAll(ctx context.Context) ([]*entity.Service, error) {
	var models []ServiceModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	services := make([]*entity.Service, len(models))
	for i, model := range models {
		var endpoints []entity.Endpoint
		if err := json.Unmarshal([]byte(model.Endpoints), &endpoints); err != nil {
			return nil, fmt.Errorf("failed to unmarshal endpoints: %w", err)
		}

		services[i] = &entity.Service{
			ID:        fmt.Sprintf("%d", model.ID),
			Name:      model.Name,
			BaseURL:   model.BaseURL,
			Endpoints: endpoints,
		}
	}

	return services, nil
}

// FindByName finds a service by name
func (r *ServiceRepository) FindByName(ctx context.Context, name string) (*entity.Service, error) {
	var model ServiceModel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find service by name: %w", err)
	}

	var endpoints []entity.Endpoint
	if err := json.Unmarshal([]byte(model.Endpoints), &endpoints); err != nil {
		return nil, fmt.Errorf("failed to unmarshal endpoints: %w", err)
	}

	service := &entity.Service{
		ID:        fmt.Sprintf("%d", model.ID),
		Name:      model.Name,
		BaseURL:   model.BaseURL,
		Endpoints: endpoints,
	}

	// Update cache
	if err := r.cache.Set(ctx, r.getCacheKey(service.ID), service, 24*time.Hour); err != nil {
		fmt.Printf("failed to cache service: %v\n", err)
	}

	return service, nil
}

// GetByID retrieves a service by its ID
func (r *ServiceRepository) GetByID(ctx context.Context, id string) (*entity.Service, error) {
	var model ServiceModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	var endpoints []entity.Endpoint
	if err := json.Unmarshal([]byte(model.Endpoints), &endpoints); err != nil {
		return nil, fmt.Errorf("failed to unmarshal endpoints: %w", err)
	}

	return &entity.Service{
		ID:        fmt.Sprintf("%d", model.ID),
		Name:      model.Name,
		BaseURL:   model.BaseURL,
		Endpoints: endpoints,
	}, nil
}

// GetByEndpoint finds services by endpoint path and method
func (r *ServiceRepository) GetByEndpoint(ctx context.Context, path string, method string) ([]*entity.Service, error) {
	var services []*entity.Service

	// Try to get from cache first
	cacheKey := fmt.Sprintf("service:endpoint:%s:%s", path, method)
	var cachedData string
	if err := r.cache.Get(ctx, cacheKey, &cachedData); err == nil {
		if err := json.Unmarshal([]byte(cachedData), &services); err == nil {
			return services, nil
		}
	}

	// Query the database
	var dbServices []ServiceModel
	if err := r.db.WithContext(ctx).Find(&dbServices).Error; err != nil {
		return nil, err
	}

	// Convert and filter by method
	for _, dbService := range dbServices {
		var endpoints []entity.Endpoint
		if err := json.Unmarshal([]byte(dbService.Endpoints), &endpoints); err != nil {
			continue // Skip this service if endpoints can't be unmarshaled
		}

		service := &entity.Service{
			ID:        fmt.Sprintf("%d", dbService.ID),
			Name:      dbService.Name,
			BaseURL:   dbService.BaseURL,
			Endpoints: endpoints,
		}

		for _, endpoint := range endpoints {
			if endpoint.Path == path {
				for _, supportedMethod := range endpoint.Methods {
					if supportedMethod == method || supportedMethod == "*" {
						services = append(services, service)
						break
					}
				}
			}
		}
	}

	if len(services) == 0 {
		return nil, errors.ErrNotFound
	}

	// Cache the result
	if jsonData, err := json.Marshal(services); err == nil {
		r.cache.Set(ctx, cacheKey, string(jsonData), time.Hour)
	}

	return services, nil
}
