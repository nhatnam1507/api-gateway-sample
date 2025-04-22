package usecase

import (
	"context"
	"testing"

	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/internal/domain/repository/mock"
	"api-gateway-sample/pkg/errors"
)

// MockLogger is a simple mock implementation of the logger.Logger interface
type MockLogger struct{}

// Debug is a no-op implementation for testing
func (l *MockLogger) Debug(msg string, args ...interface{}) {}

// Info is a no-op implementation for testing
func (l *MockLogger) Info(msg string, args ...interface{}) {}

// Warn is a no-op implementation for testing
func (l *MockLogger) Warn(msg string, args ...interface{}) {}

// Error is a no-op implementation for testing
func (l *MockLogger) Error(msg string, args ...interface{}) {}

// Fatal is a no-op implementation for testing
func (l *MockLogger) Fatal(msg string, args ...interface{}) {}

func TestServiceManagementUseCase_GetAllServices(t *testing.T) {
	// Create mock repository
	repo := mock.NewServiceRepositoryMock()
	mockLogger := &MockLogger{}

	// Create use case
	useCase := NewServiceManagementUseCase(repo, mockLogger)

	// Create test services
	service1 := &entity.Service{
		ID:          "1",
		Name:        "service1",
		Version:     "1.0.0",
		Description: "Test service 1",
		BaseURL:     "http://localhost:8081",
		Timeout:     30,
		RetryCount:  3,
		IsActive:    true,
	}

	service2 := &entity.Service{
		ID:          "2",
		Name:        "service2",
		Version:     "1.0.0",
		Description: "Test service 2",
		BaseURL:     "http://localhost:8082",
		Timeout:     30,
		RetryCount:  3,
		IsActive:    true,
	}

	// Add services to repository
	err := repo.Create(context.Background(), service1)
	if err != nil {
		t.Fatalf("Failed to create service1: %v", err)
	}

	err = repo.Create(context.Background(), service2)
	if err != nil {
		t.Fatalf("Failed to create service2: %v", err)
	}

	// Test GetAllServices
	services, err := useCase.GetAllServices(context.Background())
	if err != nil {
		t.Fatalf("GetAllServices failed: %v", err)
	}

	// Verify results
	if len(services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(services))
	}

	// Verify service IDs
	foundService1 := false
	foundService2 := false
	for _, service := range services {
		if service.ID == "1" {
			foundService1 = true
		}
		if service.ID == "2" {
			foundService2 = true
		}
	}

	if !foundService1 {
		t.Error("Service1 not found in results")
	}
	if !foundService2 {
		t.Error("Service2 not found in results")
	}
}

func TestServiceManagementUseCase_GetServiceByID(t *testing.T) {
	// Create mock repository
	repo := mock.NewServiceRepositoryMock()
	mockLogger := &MockLogger{}

	// Create use case
	useCase := NewServiceManagementUseCase(repo, mockLogger)

	// Create test service
	service := &entity.Service{
		ID:          "test-id",
		Name:        "test-service",
		Version:     "1.0.0",
		Description: "Test service",
		BaseURL:     "http://localhost:8080",
		Timeout:     30,
		RetryCount:  3,
		IsActive:    true,
	}

	// Add service to repository
	err := repo.Create(context.Background(), service)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Test GetServiceByID with valid ID
	retrievedService, err := useCase.GetServiceByID(context.Background(), "test-id")
	if err != nil {
		t.Fatalf("GetServiceByID failed: %v", err)
	}

	// Verify service properties
	if retrievedService.ID != service.ID {
		t.Errorf("Expected service ID %s, got %s", service.ID, retrievedService.ID)
	}
	if retrievedService.Name != service.Name {
		t.Errorf("Expected service name %s, got %s", service.Name, retrievedService.Name)
	}

	// Test GetServiceByID with invalid ID
	_, err = useCase.GetServiceByID(context.Background(), "non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent ID, got nil")
	}
	if !errors.IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestServiceManagementUseCase_CreateService(t *testing.T) {
	// Create mock repository
	repo := mock.NewServiceRepositoryMock()
	mockLogger := &MockLogger{}

	// Create use case
	useCase := NewServiceManagementUseCase(repo, mockLogger)

	// Create test service
	service := &entity.Service{
		Name:        "new-service",
		Version:     "1.0.0",
		Description: "New service",
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

	// Test CreateService
	err := useCase.CreateService(context.Background(), service)
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	// Verify service was created
	retrievedService, err := repo.FindByName(context.Background(), "new-service")
	if err != nil {
		t.Fatalf("Failed to find created service: %v", err)
	}

	if retrievedService.Name != service.Name {
		t.Errorf("Expected service name %s, got %s", service.Name, retrievedService.Name)
	}

	// Test creating a service with duplicate name
	duplicateService := &entity.Service{
		Name:    "new-service",
		BaseURL: "http://localhost:8081",
	}

	err = useCase.CreateService(context.Background(), duplicateService)
	if err == nil {
		t.Error("Expected error for duplicate service name, got nil")
	}
}

func TestServiceManagementUseCase_UpdateService(t *testing.T) {
	// Create mock repository
	repo := mock.NewServiceRepositoryMock()
	mockLogger := &MockLogger{}

	// Create use case
	useCase := NewServiceManagementUseCase(repo, mockLogger)

	// Create test service
	service := &entity.Service{
		ID:          "update-test-id",
		Name:        "update-test-service",
		Version:     "1.0.0",
		Description: "Test service for update",
		BaseURL:     "http://localhost:8080",
		Timeout:     30,
		RetryCount:  3,
		IsActive:    true,
	}

	// Add service to repository
	err := repo.Create(context.Background(), service)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Update service
	service.Description = "Updated description"
	service.Timeout = 60

	// Test UpdateService
	err = useCase.UpdateService(context.Background(), service)
	if err != nil {
		t.Fatalf("UpdateService failed: %v", err)
	}

	// Verify service was updated
	retrievedService, err := repo.GetByID(context.Background(), "update-test-id")
	if err != nil {
		t.Fatalf("Failed to get updated service: %v", err)
	}

	if retrievedService.Description != "Updated description" {
		t.Errorf("Expected updated description, got %s", retrievedService.Description)
	}
	if retrievedService.Timeout != 60 {
		t.Errorf("Expected updated timeout 60, got %d", retrievedService.Timeout)
	}

	// Test updating a non-existent service
	nonExistentService := &entity.Service{
		ID:   "non-existent-id",
		Name: "non-existent-service",
	}

	err = useCase.UpdateService(context.Background(), nonExistentService)
	if err == nil {
		t.Error("Expected error for non-existent service, got nil")
	}
	if !errors.IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestServiceManagementUseCase_DeleteService(t *testing.T) {
	// Create mock repository
	repo := mock.NewServiceRepositoryMock()
	mockLogger := &MockLogger{}

	// Create use case
	useCase := NewServiceManagementUseCase(repo, mockLogger)

	// Create test service
	service := &entity.Service{
		ID:          "delete-test-id",
		Name:        "delete-test-service",
		Version:     "1.0.0",
		Description: "Test service for delete",
		BaseURL:     "http://localhost:8080",
		Timeout:     30,
		RetryCount:  3,
		IsActive:    true,
	}

	// Add service to repository
	err := repo.Create(context.Background(), service)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Test DeleteService
	err = useCase.DeleteService(context.Background(), "delete-test-id")
	if err != nil {
		t.Fatalf("DeleteService failed: %v", err)
	}

	// Verify service was deleted
	_, err = repo.GetByID(context.Background(), "delete-test-id")
	if err == nil {
		t.Error("Expected error for deleted service, got nil")
	}
	if !errors.IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}

	// Test deleting a non-existent service
	err = useCase.DeleteService(context.Background(), "non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent service, got nil")
	}
	if !errors.IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}
