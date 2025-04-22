package mock

import (
	"context"
	"testing"

	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/pkg/errors"
)

func TestServiceRepositoryMock(t *testing.T) {
	// Create a new mock repository
	repo := NewServiceRepositoryMock()

	// Create a test service
	service := &entity.Service{
		ID:          "test-id",
		Name:        "test-service",
		Version:     "1.0.0",
		Description: "Test service",
		BaseURL:     "http://localhost:8080",
		Timeout:     30,
		RetryCount:  3,
		IsActive:    true,
		Endpoints: []entity.Endpoint{
			{
				Path:         "/api/test",
				Methods:      []string{"GET", "POST"},
				RateLimit:    100,
				AuthRequired: true,
			},
		},
	}

	// Test Create
	t.Run("Create", func(t *testing.T) {
		err := repo.Create(context.Background(), service)
		if err != nil {
			t.Errorf("Failed to create service: %v", err)
		}

		// Try to create a service with the same name
		duplicateService := &entity.Service{
			Name:    "test-service",
			BaseURL: "http://localhost:8081",
		}
		err = repo.Create(context.Background(), duplicateService)
		if err != errors.ErrAlreadyExists {
			t.Errorf("Expected ErrAlreadyExists, got %v", err)
		}
	})

	// Test Get
	t.Run("Get", func(t *testing.T) {
		retrievedService, err := repo.Get(context.Background(), "test-id")
		if err != nil {
			t.Errorf("Failed to get service: %v", err)
		}
		if retrievedService.Name != service.Name {
			t.Errorf("Expected service name %s, got %s", service.Name, retrievedService.Name)
		}

		// Try to get a non-existent service
		_, err = repo.Get(context.Background(), "non-existent-id")
		if err != errors.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	// Test GetByID (alias for Get)
	t.Run("GetByID", func(t *testing.T) {
		retrievedService, err := repo.GetByID(context.Background(), "test-id")
		if err != nil {
			t.Errorf("Failed to get service by ID: %v", err)
		}
		if retrievedService.Name != service.Name {
			t.Errorf("Expected service name %s, got %s", service.Name, retrievedService.Name)
		}
	})

	// Test Update
	t.Run("Update", func(t *testing.T) {
		// Update the service
		service.Description = "Updated description"
		err := repo.Update(context.Background(), service)
		if err != nil {
			t.Errorf("Failed to update service: %v", err)
		}

		// Verify the update
		retrievedService, err := repo.Get(context.Background(), "test-id")
		if err != nil {
			t.Errorf("Failed to get updated service: %v", err)
		}
		if retrievedService.Description != "Updated description" {
			t.Errorf("Expected updated description, got %s", retrievedService.Description)
		}

		// Try to update a non-existent service
		nonExistentService := &entity.Service{
			ID:   "non-existent-id",
			Name: "non-existent-service",
		}
		err = repo.Update(context.Background(), nonExistentService)
		if err != errors.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	// Test GetAll
	t.Run("GetAll", func(t *testing.T) {
		// Add another service
		anotherService := &entity.Service{
			ID:          "another-id",
			Name:        "another-service",
			Version:     "1.0.0",
			Description: "Another test service",
			BaseURL:     "http://localhost:8081",
			Timeout:     30,
			RetryCount:  3,
			IsActive:    true,
		}
		err := repo.Create(context.Background(), anotherService)
		if err != nil {
			t.Errorf("Failed to create another service: %v", err)
		}

		// Get all services
		services, err := repo.GetAll(context.Background())
		if err != nil {
			t.Errorf("Failed to get all services: %v", err)
		}
		if len(services) != 2 {
			t.Errorf("Expected 2 services, got %d", len(services))
		}
	})

	// Test FindByName
	t.Run("FindByName", func(t *testing.T) {
		retrievedService, err := repo.FindByName(context.Background(), "test-service")
		if err != nil {
			t.Errorf("Failed to find service by name: %v", err)
		}
		if retrievedService.ID != "test-id" {
			t.Errorf("Expected service ID test-id, got %s", retrievedService.ID)
		}

		// Try to find a non-existent service
		_, err = repo.FindByName(context.Background(), "non-existent-service")
		if err != errors.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	// Test GetByEndpoint
	t.Run("GetByEndpoint", func(t *testing.T) {
		services, err := repo.GetByEndpoint(context.Background(), "/api/test", "GET")
		if err != nil {
			t.Errorf("Failed to get services by endpoint: %v", err)
		}
		if len(services) != 1 {
			t.Errorf("Expected 1 service, got %d", len(services))
		}
		if services[0].ID != "test-id" {
			t.Errorf("Expected service ID test-id, got %s", services[0].ID)
		}

		// Try to get services by non-existent endpoint
		_, err = repo.GetByEndpoint(context.Background(), "/non-existent", "GET")
		if err != errors.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(context.Background(), "test-id")
		if err != nil {
			t.Errorf("Failed to delete service: %v", err)
		}

		// Verify the service is deleted
		_, err = repo.Get(context.Background(), "test-id")
		if err != errors.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}

		// Try to delete a non-existent service
		err = repo.Delete(context.Background(), "non-existent-id")
		if err != errors.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	// Test Reset
	t.Run("Reset", func(t *testing.T) {
		// Reset the repository
		repo.(*ServiceRepositoryMock).Reset()

		// Verify all services are deleted
		services, err := repo.GetAll(context.Background())
		if err != nil {
			t.Errorf("Failed to get all services after reset: %v", err)
		}
		if len(services) != 0 {
			t.Errorf("Expected 0 services after reset, got %d", len(services))
		}
	})
}
