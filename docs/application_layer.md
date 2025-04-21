# Application Layer Implementation

This document outlines the implementation details for the application layer of the API Gateway following Clean Architecture principles.

## Overview

The application layer sits between the domain layer and the infrastructure/interfaces layers. It contains use cases that orchestrate the flow of data to and from the domain entities, and applies application-specific business rules. This layer depends on the domain layer but has no dependencies on outer layers.

## Core Components

### Data Transfer Objects (DTOs)

DTOs are used to transfer data between the application layer and the interfaces layer.

#### `request_dto.go`

```go
package dto

import (
	"time"

	"api-gateway/internal/domain/entity"
)

// RequestDTO represents a client request data transfer object
type RequestDTO struct {
	ID          string              `json:"id"`
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	Headers     map[string][]string `json:"headers"`
	QueryParams map[string][]string `json:"query_params"`
	Body        []byte              `json:"body"`
	ClientIP    string              `json:"client_ip"`
	Timestamp   time.Time           `json:"timestamp"`
	Timeout     int                 `json:"timeout"` // in milliseconds
}

// ToEntity converts RequestDTO to domain entity
func (dto *RequestDTO) ToEntity() *entity.Request {
	request := entity.NewRequest(
		dto.Method,
		dto.Path,
		dto.Headers,
		dto.QueryParams,
		dto.Body,
		dto.ClientIP,
	)
	
	if dto.Timeout > 0 {
		request.SetTimeout(time.Duration(dto.Timeout) * time.Millisecond)
	}
	
	return request
}

// FromEntity creates RequestDTO from domain entity
func FromEntity(request *entity.Request) *RequestDTO {
	return &RequestDTO{
		ID:          request.ID,
		Method:      request.Method,
		Path:        request.Path,
		Headers:     request.Headers,
		QueryParams: request.QueryParams,
		Body:        request.Body,
		ClientIP:    request.ClientIP,
		Timestamp:   request.Timestamp,
		Timeout:     int(request.Timeout.Milliseconds()),
	}
}
```

#### `response_dto.go`

```go
package dto

import (
	"time"

	"api-gateway/internal/domain/entity"
)

// ResponseDTO represents a response data transfer object
type ResponseDTO struct {
	RequestID     string              `json:"request_id"`
	StatusCode    int                 `json:"status_code"`
	Headers       map[string][]string `json:"headers"`
	Body          []byte              `json:"body"`
	ContentType   string              `json:"content_type"`
	ContentLength int                 `json:"content_length"`
	Timestamp     time.Time           `json:"timestamp"`
	LatencyMs     int64               `json:"latency_ms"`
	CachedResult  bool                `json:"cached_result"`
}

// ToEntity converts ResponseDTO to domain entity
func (dto *ResponseDTO) ToEntity() *entity.Response {
	response := entity.NewResponse(
		dto.RequestID,
		dto.StatusCode,
		dto.Headers,
		dto.Body,
	)
	
	response.SetCached(dto.CachedResult)
	
	return response
}

// FromEntity creates ResponseDTO from domain entity
func FromEntity(response *entity.Response) *ResponseDTO {
	return &ResponseDTO{
		RequestID:     response.RequestID,
		StatusCode:    response.StatusCode,
		Headers:       response.Headers,
		Body:          response.Body,
		ContentType:   response.ContentType,
		ContentLength: response.ContentLength,
		Timestamp:     response.Timestamp,
		LatencyMs:     response.LatencyMs,
		CachedResult:  response.CachedResult,
	}
}
```

#### `service_dto.go`

```go
package dto

import (
	"api-gateway/internal/domain/entity"
)

// ServiceDTO represents a service data transfer object
type ServiceDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	BaseURL     string            `json:"base_url"`
	Timeout     int               `json:"timeout"`      // in milliseconds
	RetryCount  int               `json:"retry_count"`
	IsActive    bool              `json:"is_active"`
	Endpoints   []EndpointDTO     `json:"endpoints"`
	Metadata    map[string]string `json:"metadata"`
}

// EndpointDTO represents an endpoint data transfer object
type EndpointDTO struct {
	Path         string   `json:"path"`
	Methods      []string `json:"methods"`
	RateLimit    int      `json:"rate_limit"`    // requests per minute
	AuthRequired bool     `json:"auth_required"`
	Timeout      int      `json:"timeout"`       // in milliseconds
	CacheTTL     int      `json:"cache_ttl"`     // in seconds
}

// ToEntity converts ServiceDTO to domain entity
func (dto *ServiceDTO) ToEntity() *entity.Service {
	service := entity.NewService(
		dto.ID,
		dto.Name,
		dto.Version,
		dto.Description,
		dto.BaseURL,
	)
	
	service.Timeout = dto.Timeout
	service.RetryCount = dto.RetryCount
	service.IsActive = dto.IsActive
	
	for key, value := range dto.Metadata {
		service.AddMetadata(key, value)
	}
	
	for _, endpointDTO := range dto.Endpoints {
		endpoint := entity.Endpoint{
			Path:         endpointDTO.Path,
			Methods:      endpointDTO.Methods,
			RateLimit:    endpointDTO.RateLimit,
			AuthRequired: endpointDTO.AuthRequired,
			Timeout:      endpointDTO.Timeout,
			CacheTTL:     endpointDTO.CacheTTL,
		}
		service.AddEndpoint(endpoint)
	}
	
	return service
}

// FromEntity creates ServiceDTO from domain entity
func ServiceFromEntity(service *entity.Service) *ServiceDTO {
	endpointDTOs := make([]EndpointDTO, len(service.Endpoints))
	
	for i, endpoint := range service.Endpoints {
		endpointDTOs[i] = EndpointDTO{
			Path:         endpoint.Path,
			Methods:      endpoint.Methods,
			RateLimit:    endpoint.RateLimit,
			AuthRequired: endpoint.AuthRequired,
			Timeout:      endpoint.Timeout,
			CacheTTL:     endpoint.CacheTTL,
		}
	}
	
	return &ServiceDTO{
		ID:          service.ID,
		Name:        service.Name,
		Version:     service.Version,
		Description: service.Description,
		BaseURL:     service.BaseURL,
		Timeout:     service.Timeout,
		RetryCount:  service.RetryCount,
		IsActive:    service.IsActive,
		Endpoints:   endpointDTOs,
		Metadata:    service.Metadata,
	}
}
```

### Use Cases

Use cases implement the application-specific business logic and orchestrate the flow of data to and from entities.

#### `proxy_usecase.go`

```go
package usecase

import (
	"context"
	"fmt"
	"time"

	"api-gateway/internal/application/dto"
	"api-gateway/internal/domain/entity"
	"api-gateway/internal/domain/repository"
	"api-gateway/internal/domain/service"
)

// ProxyUseCase implements the use case for proxying requests to backend services
type ProxyUseCase struct {
	serviceRepo    repository.ServiceRepository
	gatewayService service.GatewayService
	authService    service.AuthService
	rateLimitService service.RateLimitService
	cacheService   service.CacheService
	logger         Logger
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewProxyUseCase creates a new ProxyUseCase
func NewProxyUseCase(
	serviceRepo repository.ServiceRepository,
	gatewayService service.GatewayService,
	authService service.AuthService,
	rateLimitService service.RateLimitService,
	cacheService service.CacheService,
	logger Logger,
) *ProxyUseCase {
	return &ProxyUseCase{
		serviceRepo:    serviceRepo,
		gatewayService: gatewayService,
		authService:    authService,
		rateLimitService: rateLimitService,
		cacheService:   cacheService,
		logger:         logger,
	}
}

// ProxyRequest proxies a request to a backend service
func (uc *ProxyUseCase) ProxyRequest(ctx context.Context, requestDTO *dto.RequestDTO) (*dto.ResponseDTO, error) {
	startTime := time.Now()
	
	// Convert DTO to domain entity
	request := requestDTO.ToEntity()
	
	// Log request
	uc.logger.Info("Received request",
		"request_id", request.ID,
		"method", request.Method,
		"path", request.Path,
		"client_ip", request.ClientIP,
	)
	
	// Validate request
	if err := uc.gatewayService.ValidateRequest(ctx, request); err != nil {
		uc.logger.Error("Request validation failed", "error", err)
		return uc.handleError(ctx, err, request)
	}
	
	// Find service for the request
	service, endpoint, err := uc.findService(ctx, request)
	if err != nil {
		uc.logger.Error("Service discovery failed", "error", err)
		return uc.handleError(ctx, err, request)
	}
	
	// Check if authentication is required
	if endpoint.AuthRequired {
		authenticated, userID, err := uc.authService.Authenticate(ctx, request)
		if err != nil {
			uc.logger.Error("Authentication failed", "error", err)
			return uc.handleError(ctx, err, request)
		}
		
		if !authenticated {
			uc.logger.Error("Unauthorized request", "request_id", request.ID)
			return uc.handleError(ctx, fmt.Errorf("unauthorized"), request)
		}
		
		request.SetAuthenticated(authenticated, userID)
		
		// Authorize the request
		if err := uc.authService.Authorize(ctx, request, service, endpoint); err != nil {
			uc.logger.Error("Authorization failed", "error", err)
			return uc.handleError(ctx, err, request)
		}
	}
	
	// Check rate limit
	if endpoint.RateLimit > 0 {
		allowed, err := uc.rateLimitService.CheckLimit(ctx, request, service, endpoint)
		if err != nil {
			uc.logger.Error("Rate limit check failed", "error", err)
			return uc.handleError(ctx, err, request)
		}
		
		if !allowed {
			uc.logger.Error("Rate limit exceeded", "request_id", request.ID)
			return uc.handleError(ctx, fmt.Errorf("rate limit exceeded"), request)
		}
		
		// Record the request for rate limiting
		if err := uc.rateLimitService.RecordRequest(ctx, request, service, endpoint); err != nil {
			uc.logger.Error("Failed to record request for rate limiting", "error", err)
			// Continue processing even if recording fails
		}
	}
	
	// Check cache
	if endpoint.CacheTTL > 0 {
		cachedResponse, found, err := uc.cacheService.Get(ctx, request)
		if err != nil {
			uc.logger.Error("Cache retrieval failed", "error", err)
			// Continue processing even if cache retrieval fails
		}
		
		if found && cachedResponse != nil {
			uc.logger.Info("Cache hit", "request_id", request.ID)
			cachedResponse.SetLatency(startTime)
			return dto.FromEntity(cachedResponse), nil
		}
	}
	
	// Transform request
	transformedRequest, err := uc.gatewayService.TransformRequest(ctx, request, service)
	if err != nil {
		uc.logger.Error("Request transformation failed", "error", err)
		return uc.handleError(ctx, err, request)
	}
	
	// Route request to backend service
	response, err := uc.gatewayService.RouteRequest(ctx, transformedRequest)
	if err != nil {
		uc.logger.Error("Request routing failed", "error", err)
		return uc.handleError(ctx, err, request)
	}
	
	// Transform response
	transformedResponse, err := uc.gatewayService.TransformResponse(ctx, response, service)
	if err != nil {
		uc.logger.Error("Response transformation failed", "error", err)
		return uc.handleError(ctx, err, request)
	}
	
	// Cache response if needed
	if endpoint.CacheTTL > 0 {
		if err := uc.cacheService.Set(ctx, request, transformedResponse, time.Duration(endpoint.CacheTTL)*time.Second); err != nil {
			uc.logger.Error("Cache storage failed", "error", err)
			// Continue processing even if cache storage fails
		}
	}
	
	// Log response
	uc.logger.Info("Sent response",
		"request_id", request.ID,
		"status_code", transformedResponse.StatusCode,
		"latency_ms", transformedResponse.LatencyMs,
	)
	
	// Return response DTO
	return dto.FromEntity(transformedResponse), nil
}

// findService finds a service for the request
func (uc *ProxyUseCase) findService(ctx context.Context, request *entity.Request) (*entity.Service, *entity.Endpoint, error) {
	services, err := uc.serviceRepo.GetByEndpoint(ctx, request.Path, request.Method)
	if err != nil {
		return nil, nil, err
	}
	
	if len(services) == 0 {
		return nil, nil, fmt.Errorf("no service found for path: %s, method: %s", request.Path, request.Method)
	}
	
	// Select the first active service
	for _, service := range services {
		if service.IsActive {
			// Find matching endpoint
			for _, endpoint := range service.Endpoints {
				if matchesEndpoint(request.Path, request.Method, endpoint) {
					return service, &endpoint, nil
				}
			}
		}
	}
	
	return nil, nil, fmt.Errorf("no active service found for path: %s, method: %s", request.Path, request.Method)
}

// matchesEndpoint checks if the request matches the endpoint
func matchesEndpoint(path, method string, endpoint entity.Endpoint) bool {
	// Check method
	methodMatches := false
	for _, m := range endpoint.Methods {
		if m == method || m == "*" {
			methodMatches = true
			break
		}
	}
	
	if !methodMatches {
		return false
	}
	
	// Check path
	// This is a simplified implementation
	// In a real application, you would use path matching with patterns and parameters
	return path == endpoint.Path || endpoint.Path == "*"
}

// handleError handles errors during request processing
func (uc *ProxyUseCase) handleError(ctx context.Context, err error, request *entity.Request) (*dto.ResponseDTO, error) {
	response, err := uc.gatewayService.HandleError(ctx, err, request)
	if err != nil {
		// If error handling fails, return a generic error response
		return &dto.ResponseDTO{
			RequestID:  request.ID,
			StatusCode: 500,
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
			Body:       []byte(`{"error":"Internal Server Error"}`),
			Timestamp:  time.Now(),
		}, nil
	}
	
	return dto.FromEntity(response), nil
}
```

#### `auth_usecase.go`

```go
package usecase

import (
	"context"
	"time"

	"api-gateway/internal/application/dto"
	"api-gateway/internal/domain/service"
)

// AuthUseCase implements the use case for authentication
type AuthUseCase struct {
	authService service.AuthService
	logger      Logger
}

// NewAuthUseCase creates a new AuthUseCase
func NewAuthUseCase(authService service.AuthService, logger Logger) *AuthUseCase {
	return &AuthUseCase{
		authService: authService,
		logger:      logger,
	}
}

// AuthRequest represents an authentication request
type AuthRequest struct {
	ClientID     string            `json:"client_id"`
	ClientSecret string            `json:"client_secret"`
	GrantType    string            `json:"grant_type"`
	Scope        string            `json:"scope"`
	Username     string            `json:"username,omitempty"`
	Password     string            `json:"password,omitempty"`
	RefreshToken string            `json:"refresh_token,omitempty"`
	Claims       map[string]string `json:"claims,omitempty"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// Authenticate authenticates a client and returns an access token
func (uc *AuthUseCase) Authenticate(ctx context.Context, request *AuthRequest) (*AuthResponse, error) {
	uc.logger.Info("Authentication request",
		"client_id", request.ClientID,
		"grant_type", request.GrantType,
		"scope", request.Scope,
	)
	
	// Validate request
	if err := uc.validateAuthRequest(request); err != nil {
		uc.logger.Error("Authentication request validation failed", "error", err)
		return nil, err
	}
	
	// Process based on grant type
	var userID string
	var err error
	
	switch request.GrantType {
	case "client_credentials":
		userID = request.ClientID
	case "password":
		// Authenticate with username and password
		// Implementation details
		userID = request.Username
	case "refresh_token":
		// Validate refresh token
		// Implementation details
		claims, err := uc.authService.ValidateToken(ctx, request.RefreshToken)
		if err != nil {
			uc.logger.Error("Refresh token validation failed", "error", err)
			return nil, err
		}
		userID = claims["sub"].(string)
	default:
		uc.logger.Error("Unsupported grant type", "grant_type", request.GrantType)
		return nil, ErrUnsupportedGrantType
	}
	
	// Prepare claims
	claims := make(map[string]interface{})
	claims["sub"] = userID
	claims["scope"] = request.Scope
	
	for key, value := range request.Claims {
		claims[key] = value
	}
	
	// Generate access token
	accessToken, err := uc.authService.GenerateToken(ctx, userID, claims)
	if err != nil {
		uc.logger.Error("Token generation failed", "error", err)
		return nil, err
	}
	
	// Generate refresh token if needed
	var refreshToken string
	if request.GrantType != "refresh_token" {
		refreshClaims := map[string]interface{}{
			"sub":  userID,
			"type": "refresh",
		}
		refreshToken, err = uc.authService.GenerateToken(ctx, userID, refreshClaims)
		if err != nil {
			uc.logger.Error("Refresh token generation failed", "error", err)
			// Continue even if refresh token generation fails
		}
	} else {
		refreshToken = request.RefreshToken
	}
	
	uc.logger.Info("Authentication successful", "user_id", userID)
	
	// Return response
	return &AuthResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		RefreshToken: refreshToken,
		Scope:        request.Scope,
	}, nil
}

// ValidateToken validates a token and returns its claims
func (uc *AuthUseCase) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	uc.logger.Info("Token validation request")
	
	// Validate token
	claims, err := uc.authService.ValidateToken(ctx, token)
	if err != nil {
		uc.logger.Error("Token validation failed", "error", err)
		return nil, err
	}
	
	uc.logger.Info("Token validation successful", "user_id", claims["sub"])
	
	return claims, nil
}

// validateAuthRequest validates an authentication request
func (uc *AuthUseCase) validateAuthRequest(request *AuthRequest) error {
	if request.ClientID == "" {
		return ErrMissingClientID
	}
	
	switch request.GrantType {
	case "client_credentials":
		if request.ClientSecret == "" {
			return ErrMissingClientSecret
		}
	case "password":
		if request.Username == "" {
			return ErrMissingUsername
		}
		if request.Password == "" {
			return ErrMissingPassword
		}
	case "refresh_token":
		if request.RefreshToken == "" {
			return ErrMissingRefreshToken
		}
	default:
		return ErrUnsupportedGrantType
	}
	
	return nil
}

// Error definitions
var (
	ErrMissingClientID      = &AuthError{Code: "missing_client_id", Message: "Client ID is required"}
	ErrMissingClientSecret  = &AuthError{Code: "missing_client_secret", Message: "Client secret is required"}
	ErrMissingUsername      = &AuthError{Code: "missing_username", Message: "Username is required"}
	ErrMissingPassword      = &AuthError{Code: "missing_password", Message: "Password is required"}
	ErrMissingRefreshToken  = &AuthError{Code: "missing_refresh_token", Message: "Refresh token is required"}
	ErrUnsupportedGrantType = &AuthError{Code: "unsupported_grant_type", Message: "Unsupported grant type"}
)

// AuthError represents an authentication error
type AuthError struct {
	Code    string `json:"error"`
	Message string `json:"error_description"`
}

// Error returns the error message
func (e *AuthError) Error() string {
	return e.Message
}
```

#### `rate_limit_usecase.go`

```go
package usecase

import (
	"context"
	"fmt"
	"time"

	"api-gateway/internal/domain/entity"
	"api-gateway/internal/domain/service"
)

// RateLimitUseCase implements the use case for rate limiting
type RateLimitUseCase struct {
	rateLimitService service.RateLimitService
	logger           Logger
}

// NewRateLimitUseCase creates a new RateLimitUseCase
func NewRateLimitUseCase(rateLimitService service.RateLimitService, logger Logger) *RateLimitUseCase {
	return &RateLimitUseCase{
		rateLimitService: rateLimitService,
		logger:           logger,
	}
}

// RateLimitInfo represents rate limit information
type RateLimitInfo struct {
	Limit     int   `json:"limit"`
	Remaining int   `json:"remaining"`
	Reset     int64 `json:"reset"`
}

// CheckRateLimit checks if a request exceeds the rate limit
func (uc *RateLimitUseCase) CheckRateLimit(ctx context.Context, clientID string, path, method string) (*RateLimitInfo, error) {
	uc.logger.Info("Rate limit check",
		"client_id", clientID,
		"path", path,
		"method", method,
	)
	
	// Create a dummy request for rate limit checking
	request := &entity.Request{
		Method:   method,
		Path:     path,
		ClientIP: clientID,
		UserID:   clientID,
	}
	
	// Create dummy service and endpoint for rate limit checking
	service := &entity.Service{
		ID:   "rate-limit-check",
		Name: "rate-limit-check",
	}
	
	endpoint := &entity.Endpoint{
		Path:      path,
		Methods:   []string{method},
		RateLimit: 0, // Will be determined by the rate limit service
	}
	
	// Get current limit
	limit, remaining, err := uc.rateLimitService.GetLimit(ctx, clientID, service, endpoint)
	if err != nil {
		uc.logger.Error("Failed to get rate limit", "error", err)
		return nil, err
	}
	
	// Check if limit is exceeded
	allowed, err := uc.rateLimitService.CheckLimit(ctx, request, service, endpoint)
	if err != nil {
		uc.logger.Error("Rate limit check failed", "error", err)
		return nil, err
	}
	
	if !allowed {
		uc.logger.Info("Rate limit exceeded",
			"client_id", clientID,
			"path", path,
			"method", method,
		)
		return &RateLimitInfo{
			Limit:     limit,
			Remaining: 0,
			Reset:     time.Now().Add(time.Minute).Unix(), // Reset after 1 minute
		}, fmt.Errorf("rate limit exceeded")
	}
	
	uc.logger.Info("Rate limit check passed",
		"client_id", clientID,
		"path", path,
		"method", method,
		"remaining", remaining,
	)
	
	return &RateLimitInfo{
		Limit:     limit,
		Remaining: remaining,
		Reset:     time.Now().Add(time.Minute).Unix(), // Reset after 1 minute
	}, nil
}

// UpdateRateLimit updates the rate limit for a client
func (uc *RateLimitUseCase) UpdateRateLimit(ctx context.Context, clientID string, path string, limit int) error {
	uc.logger.Info("Rate limit update",
		"client_id", clientID,
		"path", path,
		"limit", limit,
	)
	
	// Implementation details
	// This would typically involve updating a configuration in a database or cache
	
	return nil
}
```

#### `service_management_usecase.go`

```go
package usecase

import (
	"context"

	"api-gateway/internal/application/dto"
	"api-gateway/internal/domain/entity"
	"api-gateway/internal/domain/repository"
)

// ServiceManagementUseCase implements the use case for service management
type ServiceManagementUseCase struct {
	serviceRepo repository.ServiceRepository
	logger      Logger
}

// NewServiceManagementUseCase creates a new ServiceManagementUseCase
func NewServiceManagementUseCase(serviceRepo repository.ServiceRepository, logger Logger) *ServiceManagementUseCase {
	return &ServiceManagementUseCase{
		serviceRepo: serviceRepo,
		logger:      logger,
	}
}

// GetAllServices returns all registered services
func (uc *ServiceManagementUseCase) GetAllServices(ctx context.Context) ([]*dto.ServiceDTO, error) {
	uc.logger.Info("Get all services request")
	
	// Get all services from repository
	services, err := uc.serviceRepo.GetAll(ctx)
	if err != nil {
		uc.logger.Error("Failed to get services", "error", err)
		return nil, err
	}
	
	// Convert to DTOs
	serviceDTOs := make([]*dto.ServiceDTO, len(services))
	for i, service := range services {
		serviceDTOs[i] = dto.ServiceFromEntity(service)
	}
	
	uc.logger.Info("Retrieved services", "count", len(serviceDTOs))
	
	return serviceDTOs, nil
}

// GetServiceByID returns a service by its ID
func (uc *ServiceManagementUseCase) GetServiceByID(ctx context.Context, id string) (*dto.ServiceDTO, error) {
	uc.logger.Info("Get service by ID request", "id", id)
	
	// Get service from repository
	service, err := uc.serviceRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get service", "id", id, "error", err)
		return nil, err
	}
	
	// Convert to DTO
	serviceDTO := dto.ServiceFromEntity(service)
	
	uc.logger.Info("Retrieved service", "id", id, "name", service.Name)
	
	return serviceDTO, nil
}

// CreateService creates a new service
func (uc *ServiceManagementUseCase) CreateService(ctx context.Context, serviceDTO *dto.ServiceDTO) (*dto.ServiceDTO, error) {
	uc.logger.Info("Create service request", "name", serviceDTO.Name)
	
	// Convert DTO to entity
	service := serviceDTO.ToEntity()
	
	// Create service in repository
	err := uc.serviceRepo.Create(ctx, service)
	if err != nil {
		uc.logger.Error("Failed to create service", "name", serviceDTO.Name, "error", err)
		return nil, err
	}
	
	// Convert back to DTO
	createdServiceDTO := dto.ServiceFromEntity(service)
	
	uc.logger.Info("Service created", "id", service.ID, "name", service.Name)
	
	return createdServiceDTO, nil
}

// UpdateService updates an existing service
func (uc *ServiceManagementUseCase) UpdateService(ctx context.Context, serviceDTO *dto.ServiceDTO) (*dto.ServiceDTO, error) {
	uc.logger.Info("Update service request", "id", serviceDTO.ID)
	
	// Check if service exists
	_, err := uc.serviceRepo.GetByID(ctx, serviceDTO.ID)
	if err != nil {
		uc.logger.Error("Service not found", "id", serviceDTO.ID, "error", err)
		return nil, err
	}
	
	// Convert DTO to entity
	service := serviceDTO.ToEntity()
	
	// Update service in repository
	err = uc.serviceRepo.Update(ctx, service)
	if err != nil {
		uc.logger.Error("Failed to update service", "id", serviceDTO.ID, "error", err)
		return nil, err
	}
	
	// Convert back to DTO
	updatedServiceDTO := dto.ServiceFromEntity(service)
	
	uc.logger.Info("Service updated", "id", service.ID, "name", service.Name)
	
	return updatedServiceDTO, nil
}

// DeleteService deletes a service by its ID
func (uc *ServiceManagementUseCase) DeleteService(ctx context.Context, id string) error {
	uc.logger.Info("Delete service request", "id", id)
	
	// Check if service exists
	_, err := uc.serviceRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Service not found", "id", id, "error", err)
		return err
	}
	
	// Delete service from repository
	err = uc.serviceRepo.Delete(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to delete service", "id", id, "error", err)
		return err
	}
	
	uc.logger.Info("Service deleted", "id", id)
	
	return nil
}
```

### Application Services

Application services coordinate the execution of use cases and provide additional functionality.

#### `metrics_service.go`

```go
package service

import (
	"context"
	"sync"
	"time"
)

// MetricsService collects and provides metrics for the API Gateway
type MetricsService struct {
	requestCount      map[string]int64
	requestLatencies  map[string][]int64
	errorCount        map[string]int64
	statusCodeCounts  map[int]int64
	serviceLatencies  map[string][]int64
	mutex             sync.RWMutex
	metricsAggregator MetricsAggregator
}

// MetricsAggregator defines the interface for metrics aggregation
type MetricsAggregator interface {
	RecordRequest(path string, method string, statusCode int, latencyMs int64)
	RecordError(path string, method string, errorType string)
	RecordServiceLatency(serviceID string, latencyMs int64)
}

// NewMetricsService creates a new MetricsService
func NewMetricsService(metricsAggregator MetricsAggregator) *MetricsService {
	return &MetricsService{
		requestCount:      make(map[string]int64),
		requestLatencies:  make(map[string][]int64),
		errorCount:        make(map[string]int64),
		statusCodeCounts:  make(map[int]int64),
		serviceLatencies:  make(map[string][]int64),
		metricsAggregator: metricsAggregator,
	}
}

// RecordRequest records a request for metrics
func (s *MetricsService) RecordRequest(ctx context.Context, path, method string, statusCode int, latencyMs int64) {
	key := method + ":" + path
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Update request count
	s.requestCount[key]++
	
	// Update latencies
	s.requestLatencies[key] = append(s.requestLatencies[key], latencyMs)
	
	// Update status code counts
	s.statusCodeCounts[statusCode]++
	
	// Forward to metrics aggregator
	if s.metricsAggregator != nil {
		s.metricsAggregator.RecordRequest(path, method, statusCode, latencyMs)
	}
}

// RecordError records an error for metrics
func (s *MetricsService) RecordError(ctx context.Context, path, method, errorType string) {
	key := method + ":" + path
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Update error count
	s.errorCount[key]++
	
	// Forward to metrics aggregator
	if s.metricsAggregator != nil {
		s.metricsAggregator.RecordError(path, method, errorType)
	}
}

// RecordServiceLatency records service latency for metrics
func (s *MetricsService) RecordServiceLatency(ctx context.Context, serviceID string, latencyMs int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Update service latencies
	s.serviceLatencies[serviceID] = append(s.serviceLatencies[serviceID], latencyMs)
	
	// Forward to metrics aggregator
	if s.metricsAggregator != nil {
		s.metricsAggregator.RecordServiceLatency(serviceID, latencyMs)
	}
}

// GetRequestCount returns the request count for a path and method
func (s *MetricsService) GetRequestCount(ctx context.Context, path, method string) int64 {
	key := method + ":" + path
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return s.requestCount[key]
}

// GetAverageLatency returns the average latency for a path and method
func (s *MetricsService) GetAverageLatency(ctx context.Context, path, method string) float64 {
	key := method + ":" + path
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	latencies := s.requestLatencies[key]
	if len(latencies) == 0 {
		return 0
	}
	
	var sum int64
	for _, latency := range latencies {
		sum += latency
	}
	
	return float64(sum) / float64(len(latencies))
}

// GetErrorRate returns the error rate for a path and method
func (s *MetricsService) GetErrorRate(ctx context.Context, path, method string) float64 {
	key := method + ":" + path
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	requestCount := s.requestCount[key]
	if requestCount == 0 {
		return 0
	}
	
	errorCount := s.errorCount[key]
	
	return float64(errorCount) / float64(requestCount)
}

// GetStatusCodeCounts returns the status code counts
func (s *MetricsService) GetStatusCodeCounts(ctx context.Context) map[int]int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// Create a copy to avoid concurrent access issues
	counts := make(map[int]int64)
	for code, count := range s.statusCodeCounts {
		counts[code] = count
	}
	
	return counts
}

// GetServiceLatency returns the average latency for a service
func (s *MetricsService) GetServiceLatency(ctx context.Context, serviceID string) float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	latencies := s.serviceLatencies[serviceID]
	if len(latencies) == 0 {
		return 0
	}
	
	var sum int64
	for _, latency := range latencies {
		sum += latency
	}
	
	return float64(sum) / float64(len(latencies))
}

// ResetMetrics resets all metrics
func (s *MetricsService) ResetMetrics(ctx context.Context) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.requestCount = make(map[string]int64)
	s.requestLatencies = make(map[string][]int64)
	s.errorCount = make(map[string]int64)
	s.statusCodeCounts = make(map[int]int64)
	s.serviceLatencies = make(map[string][]int64)
}
```

#### `circuit_breaker_service.go`

```go
package service

import (
	"context"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	// CircuitBreakerStateClosed represents a closed circuit (allowing requests)
	CircuitBreakerStateClosed CircuitBreakerState = "CLOSED"
	
	// CircuitBreakerStateOpen represents an open circuit (blocking requests)
	CircuitBreakerStateOpen CircuitBreakerState = "OPEN"
	
	// CircuitBreakerStateHalfOpen represents a half-open circuit (allowing test requests)
	CircuitBreakerStateHalfOpen CircuitBreakerState = "HALF_OPEN"
)

// CircuitBreakerConfig represents the configuration for a circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold   int           // Number of failures before opening the circuit
	SuccessThreshold   int           // Number of successes before closing the circuit
	Timeout            time.Duration // Time to wait before transitioning from open to half-open
	MaxConcurrentCalls int           // Maximum number of concurrent calls allowed
}

// CircuitBreaker represents a circuit breaker
type CircuitBreaker struct {
	ServiceID        string
	State            CircuitBreakerState
	FailureCount     int
	SuccessCount     int
	LastStateChange  time.Time
	LastFailure      time.Time
	Config           CircuitBreakerConfig
	mutex            sync.RWMutex
	concurrentCalls  int
	concurrentMutex  sync.Mutex
}

// CircuitBreakerService manages circuit breakers for backend services
type CircuitBreakerService struct {
	circuitBreakers map[string]*CircuitBreaker
	mutex           sync.RWMutex
	logger          Logger
}

// NewCircuitBreakerService creates a new CircuitBreakerService
func NewCircuitBreakerService(logger Logger) *CircuitBreakerService {
	return &CircuitBreakerService{
		circuitBreakers: make(map[string]*CircuitBreaker),
		logger:          logger,
	}
}

// GetCircuitBreaker gets or creates a circuit breaker for a service
func (s *CircuitBreakerService) GetCircuitBreaker(ctx context.Context, serviceID string) *CircuitBreaker {
	s.mutex.RLock()
	cb, exists := s.circuitBreakers[serviceID]
	s.mutex.RUnlock()
	
	if exists {
		return cb
	}
	
	// Create new circuit breaker
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Check again in case another goroutine created it
	cb, exists = s.circuitBreakers[serviceID]
	if exists {
		return cb
	}
	
	cb = &CircuitBreaker{
		ServiceID:       serviceID,
		State:           CircuitBreakerStateClosed,
		LastStateChange: time.Now(),
		Config: CircuitBreakerConfig{
			FailureThreshold:   5,
			SuccessThreshold:   3,
			Timeout:            30 * time.Second,
			MaxConcurrentCalls: 100,
		},
	}
	
	s.circuitBreakers[serviceID] = cb
	
	return cb
}

// AllowRequest checks if a request is allowed based on the circuit breaker state
func (s *CircuitBreakerService) AllowRequest(ctx context.Context, serviceID string) bool {
	cb := s.GetCircuitBreaker(ctx, serviceID)
	
	cb.mutex.RLock()
	state := cb.State
	timeout := cb.Config.Timeout
	lastStateChange := cb.LastStateChange
	cb.mutex.RUnlock()
	
	// Check if circuit is open
	if state == CircuitBreakerStateOpen {
		// Check if timeout has elapsed
		if time.Since(lastStateChange) > timeout {
			// Transition to half-open
			cb.mutex.Lock()
			cb.State = CircuitBreakerStateHalfOpen
			cb.LastStateChange = time.Now()
			cb.mutex.Unlock()
			
			s.logger.Info("Circuit breaker transitioned to half-open", "service_id", serviceID)
			
			// Allow the request
			return s.incrementConcurrentCalls(cb)
		}
		
		// Circuit is open and timeout has not elapsed
		return false
	}
	
	// Circuit is closed or half-open
	return s.incrementConcurrentCalls(cb)
}

// incrementConcurrentCalls increments the concurrent calls counter
func (s *CircuitBreakerService) incrementConcurrentCalls(cb *CircuitBreaker) bool {
	cb.concurrentMutex.Lock()
	defer cb.concurrentMutex.Unlock()
	
	if cb.concurrentCalls >= cb.Config.MaxConcurrentCalls {
		return false
	}
	
	cb.concurrentCalls++
	return true
}

// decrementConcurrentCalls decrements the concurrent calls counter
func (s *CircuitBreakerService) decrementConcurrentCalls(cb *CircuitBreaker) {
	cb.concurrentMutex.Lock()
	defer cb.concurrentMutex.Unlock()
	
	cb.concurrentCalls--
	if cb.concurrentCalls < 0 {
		cb.concurrentCalls = 0
	}
}

// RecordSuccess records a successful request
func (s *CircuitBreakerService) RecordSuccess(ctx context.Context, serviceID string) {
	cb := s.GetCircuitBreaker(ctx, serviceID)
	
	defer s.decrementConcurrentCalls(cb)
	
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	// Reset failure count
	cb.FailureCount = 0
	
	// If circuit is half-open, increment success count
	if cb.State == CircuitBreakerStateHalfOpen {
		cb.SuccessCount++
		
		// If success threshold is reached, close the circuit
		if cb.SuccessCount >= cb.Config.SuccessThreshold {
			cb.State = CircuitBreakerStateClosed
			cb.LastStateChange = time.Now()
			cb.SuccessCount = 0
			
			s.logger.Info("Circuit breaker closed", "service_id", serviceID)
		}
	}
}

// RecordFailure records a failed request
func (s *CircuitBreakerService) RecordFailure(ctx context.Context, serviceID string) {
	cb := s.GetCircuitBreaker(ctx, serviceID)
	
	defer s.decrementConcurrentCalls(cb)
	
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	// Record failure
	cb.FailureCount++
	cb.LastFailure = time.Now()
	
	// Reset success count
	cb.SuccessCount = 0
	
	// If circuit is closed and failure threshold is reached, open the circuit
	if cb.State == CircuitBreakerStateClosed && cb.FailureCount >= cb.Config.FailureThreshold {
		cb.State = CircuitBreakerStateOpen
		cb.LastStateChange = time.Now()
		
		s.logger.Info("Circuit breaker opened", "service_id", serviceID)
	}
	
	// If circuit is half-open, open the circuit
	if cb.State == CircuitBreakerStateHalfOpen {
		cb.State = CircuitBreakerStateOpen
		cb.LastStateChange = time.Now()
		
		s.logger.Info("Circuit breaker reopened", "service_id", serviceID)
	}
}

// GetState returns the state of a circuit breaker
func (s *CircuitBreakerService) GetState(ctx context.Context, serviceID string) CircuitBreakerState {
	cb := s.GetCircuitBreaker(ctx, serviceID)
	
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	return cb.State
}

// Reset resets a circuit breaker to its initial state
func (s *CircuitBreakerService) Reset(ctx context.Context, serviceID string) {
	cb := s.GetCircuitBreaker(ctx, serviceID)
	
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.State = CircuitBreakerStateClosed
	cb.FailureCount = 0
	cb.SuccessCount = 0
	cb.LastStateChange = time.Now()
	
	s.logger.Info("Circuit breaker reset", "service_id", serviceID)
}

// UpdateConfig updates the configuration for a circuit breaker
func (s *CircuitBreakerService) UpdateConfig(ctx context.Context, serviceID string, config CircuitBreakerConfig) {
	cb := s.GetCircuitBreaker(ctx, serviceID)
	
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.Config = config
	
	s.logger.Info("Circuit breaker config updated", "service_id", serviceID)
}
```

## Conclusion

The application layer orchestrates the flow of data to and from the domain entities and applies application-specific business rules. It depends on the domain layer but has no dependencies on outer layers, maintaining the clean architecture principles.

Key components of the application layer include:

1. **Data Transfer Objects (DTOs)**: Used to transfer data between the application layer and the interfaces layer.
2. **Use Cases**: Implement the application-specific business logic and orchestrate the flow of data.
3. **Application Services**: Provide additional functionality and coordinate the execution of use cases.

The application layer is responsible for:

- Transforming data between the domain layer and the interfaces layer
- Implementing business rules that span multiple domain entities
- Coordinating the execution of domain services
- Handling cross-cutting concerns like logging, metrics, and circuit breaking

By following clean architecture principles, the application layer remains independent of infrastructure concerns, making it easier to test and maintain.
