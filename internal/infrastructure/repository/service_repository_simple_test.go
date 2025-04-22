package repository

import (
	"context"
	"testing"

	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/internal/domain/repository/mock"
)

func TestServiceRepository(t *testing.T) {
	// Use the mock repository instead of the actual implementation
	// This allows us to test the functionality without a real database
	repo := mock.NewServiceRepositoryMock()

	// Test data
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
				Timeout:      30,
			},
		},
	}

	// Test Create
	t.Run("Create", func(t *testing.T) {
		err := repo.Create(context.Background(), service)
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}
	})

	// Test Get
	t.Run("Get", func(t *testing.T) {
		retrievedService, err := repo.Get(context.Background(), "test-id")
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}
		if retrievedService.Name != service.Name {
			t.Errorf("Get() got = %v, want %v", retrievedService.Name, service.Name)
		}
	})

	// Test GetAll
	t.Run("GetAll", func(t *testing.T) {
		services, err := repo.GetAll(context.Background())
		if err != nil {
			t.Errorf("GetAll() error = %v", err)
		}
		if len(services) != 1 {
			t.Errorf("GetAll() got = %v, want %v", len(services), 1)
		}
	})

	// Test Update
	t.Run("Update", func(t *testing.T) {
		service.Description = "Updated description"
		err := repo.Update(context.Background(), service)
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}

		// Verify update
		updatedService, err := repo.Get(context.Background(), "test-id")
		if err != nil {
			t.Errorf("Get() after update error = %v", err)
		}
		if updatedService.Description != "Updated description" {
			t.Errorf("Update() got = %v, want %v", updatedService.Description, "Updated description")
		}
	})

	// Test FindByName
	t.Run("FindByName", func(t *testing.T) {
		retrievedService, err := repo.FindByName(context.Background(), "test-service")
		if err != nil {
			t.Errorf("FindByName() error = %v", err)
		}
		if retrievedService.ID != "test-id" {
			t.Errorf("FindByName() got = %v, want %v", retrievedService.ID, "test-id")
		}
	})

	// Test GetByEndpoint
	t.Run("GetByEndpoint", func(t *testing.T) {
		services, err := repo.GetByEndpoint(context.Background(), "/api/test", "GET")
		if err != nil {
			t.Errorf("GetByEndpoint() error = %v", err)
		}
		if len(services) != 1 {
			t.Errorf("GetByEndpoint() got = %v, want %v", len(services), 1)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(context.Background(), "test-id")
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		// Verify deletion
		_, err = repo.Get(context.Background(), "test-id")
		if err == nil {
			t.Error("Delete() failed, service still exists")
		}
	})
}
