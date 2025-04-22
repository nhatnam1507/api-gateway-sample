package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-gateway-sample/internal/application/dto"
	"api-gateway-sample/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockServiceUseCase is a mock implementation of the ServiceUseCase
type MockServiceUseCase struct {
	mock.Mock
}

func (m *MockServiceUseCase) CreateService(ctx context.Context, req *dto.CreateServiceRequest) (*dto.ServiceResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ServiceResponse), args.Error(1)
}

func (m *MockServiceUseCase) GetService(ctx context.Context, id string) (*dto.ServiceResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ServiceResponse), args.Error(1)
}

func (m *MockServiceUseCase) UpdateService(ctx context.Context, id string, req *dto.UpdateServiceRequest) (*dto.ServiceResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ServiceResponse), args.Error(1)
}

func (m *MockServiceUseCase) DeleteService(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockServiceUseCase) ListServices(ctx context.Context) ([]*dto.ServiceResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.ServiceResponse), args.Error(1)
}

func (m *MockServiceUseCase) FindServiceByName(ctx context.Context, name string) (*dto.ServiceResponse, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ServiceResponse), args.Error(1)
}

func TestCreateServiceSimple(t *testing.T) {
	// Create mock use case
	mockUseCase := new(MockServiceUseCase)

	// Create handler with the mock
	handler := &ServiceHandler{
		serviceUseCase: mockUseCase,
	}

	// Test data
	createReq := &dto.CreateServiceRequest{
		Name:    "test-service",
		BaseURL: "http://localhost:8080",
		Endpoints: []dto.EndpointConfig{
			{
				Path:         "/api/test",
				Methods:      []string{"GET", "POST"},
				RateLimit:    100,
				AuthRequired: true,
				Timeout:      30,
			},
		},
	}

	serviceResp := &dto.ServiceResponse{
		ID:      "test-id",
		Name:    "test-service",
		BaseURL: "http://localhost:8080",
		Endpoints: []dto.EndpointConfig{
			{
				Path:         "/api/test",
				Methods:      []string{"GET", "POST"},
				RateLimit:    100,
				AuthRequired: true,
				Timeout:      30,
			},
		},
	}

	// Set up expectations
	mockUseCase.On("CreateService", mock.Anything, mock.MatchedBy(func(req *dto.CreateServiceRequest) bool {
		return req.Name == createReq.Name
	})).Return(serviceResp, nil)

	// Create request
	reqBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, "/services", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.CreateService(rr, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, rr.Code)

	var respBody dto.ServiceResponse
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, serviceResp.ID, respBody.ID)
	assert.Equal(t, serviceResp.Name, respBody.Name)

	// Verify expectations
	mockUseCase.AssertExpectations(t)
}

func TestGetServiceSimple(t *testing.T) {
	// Create mock use case
	mockUseCase := new(MockServiceUseCase)

	// Create handler with the mock
	handler := &ServiceHandler{
		serviceUseCase: mockUseCase,
	}

	// Test data
	serviceID := "test-id"
	serviceResp := &dto.ServiceResponse{
		ID:      serviceID,
		Name:    "test-service",
		BaseURL: "http://localhost:8080",
		Endpoints: []dto.EndpointConfig{},
	}

	// Set up expectations
	mockUseCase.On("GetService", mock.Anything, serviceID).Return(serviceResp, nil)

	// Create request
	req, _ := http.NewRequest(http.MethodGet, "/services/"+serviceID, nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Set up router to extract path variables
	router := mux.NewRouter()
	router.HandleFunc("/services/{id}", handler.GetService).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var respBody dto.ServiceResponse
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, serviceResp.ID, respBody.ID)
	assert.Equal(t, serviceResp.Name, respBody.Name)

	// Verify expectations
	mockUseCase.AssertExpectations(t)
}

func TestGetServiceNotFoundSimple(t *testing.T) {
	// Create mock use case
	mockUseCase := new(MockServiceUseCase)

	// Create handler with the mock
	handler := &ServiceHandler{
		serviceUseCase: mockUseCase,
	}

	// Test data
	serviceID := "non-existent-id"

	// Set up expectations
	mockUseCase.On("GetService", mock.Anything, serviceID).Return(nil, errors.ErrNotFound)

	// Create request
	req, _ := http.NewRequest(http.MethodGet, "/services/"+serviceID, nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Set up router to extract path variables
	router := mux.NewRouter()
	router.HandleFunc("/services/{id}", handler.GetService).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusNotFound, rr.Code)

	// Verify expectations
	mockUseCase.AssertExpectations(t)
}

func TestListServicesSimple(t *testing.T) {
	// Create mock use case
	mockUseCase := new(MockServiceUseCase)

	// Create handler with the mock
	handler := &ServiceHandler{
		serviceUseCase: mockUseCase,
	}

	// Test data
	services := []*dto.ServiceResponse{
		{
			ID:      "1",
			Name:    "service1",
			BaseURL: "http://localhost:8081",
			Endpoints: []dto.EndpointConfig{},
		},
		{
			ID:      "2",
			Name:    "service2",
			BaseURL: "http://localhost:8082",
			Endpoints: []dto.EndpointConfig{},
		},
	}

	// Set up expectations
	mockUseCase.On("ListServices", mock.Anything).Return(services, nil)

	// Create request
	req, _ := http.NewRequest(http.MethodGet, "/services", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.ListServices(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	var respBody []*dto.ServiceResponse
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, len(services), len(respBody))
	assert.Equal(t, services[0].ID, respBody[0].ID)
	assert.Equal(t, services[1].ID, respBody[1].ID)

	// Verify expectations
	mockUseCase.AssertExpectations(t)
}
