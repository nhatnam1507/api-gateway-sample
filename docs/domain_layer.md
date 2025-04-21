# Domain Layer Implementation

This document outlines the implementation details for the domain layer of the API Gateway following Clean Architecture principles.

## Overview

The domain layer is the innermost layer of the Clean Architecture and contains the core business entities and business rules. It has no dependencies on other layers and defines interfaces that outer layers must implement.

## Core Components

### Entities

Entities represent the core business objects of the API Gateway.

#### `request.go`

```go
package entity

import (
	"net/http"
	"time"
)

// Request represents a client request to the API Gateway
type Request struct {
	ID            string
	Method        string
	Path          string
	Headers       map[string][]string
	QueryParams   map[string][]string
	Body          []byte
	ClientIP      string
	Timestamp     time.Time
	Authenticated bool
	UserID        string
	Timeout       time.Duration
}

// NewRequest creates a new Request entity
func NewRequest(method, path string, headers map[string][]string, queryParams map[string][]string, body []byte, clientIP string) *Request {
	return &Request{
		ID:          generateID(),
		Method:      method,
		Path:        path,
		Headers:     headers,
		QueryParams: queryParams,
		Body:        body,
		ClientIP:    clientIP,
		Timestamp:   time.Now(),
		Timeout:     30 * time.Second, // Default timeout
	}
}

// SetAuthenticated sets the authentication status and user ID
func (r *Request) SetAuthenticated(authenticated bool, userID string) {
	r.Authenticated = authenticated
	r.UserID = userID
}

// SetTimeout sets the request timeout
func (r *Request) SetTimeout(timeout time.Duration) {
	r.Timeout = timeout
}

// ToHTTPRequest converts the Request entity to an http.Request
func (r *Request) ToHTTPRequest(targetURL string) (*http.Request, error) {
	// Implementation details
	return nil, nil
}

// generateID generates a unique request ID
func generateID() string {
	// Implementation details
	return ""
}
```

#### `response.go`

```go
package entity

import (
	"net/http"
	"time"
)

// Response represents a response from a backend service
type Response struct {
	RequestID     string
	StatusCode    int
	Headers       map[string][]string
	Body          []byte
	ContentType   string
	ContentLength int
	Timestamp     time.Time
	LatencyMs     int64
	CachedResult  bool
}

// NewResponse creates a new Response entity
func NewResponse(requestID string, statusCode int, headers map[string][]string, body []byte) *Response {
	return &Response{
		RequestID:     requestID,
		StatusCode:    statusCode,
		Headers:       headers,
		Body:          body,
		ContentType:   extractContentType(headers),
		ContentLength: len(body),
		Timestamp:     time.Now(),
		CachedResult:  false,
	}
}

// SetLatency sets the latency of the request
func (r *Response) SetLatency(startTime time.Time) {
	r.LatencyMs = time.Since(startTime).Milliseconds()
}

// SetCached marks the response as cached
func (r *Response) SetCached(cached bool) {
	r.CachedResult = cached
}

// ToHTTPResponse converts the Response entity to an http.Response
func (r *Response) ToHTTPResponse() *http.Response {
	// Implementation details
	return nil
}

// extractContentType extracts the content type from headers
func extractContentType(headers map[string][]string) string {
	// Implementation details
	return ""
}
```

#### `service.go`

```go
package entity

// Service represents a backend service that can be called by the API Gateway
type Service struct {
	ID          string
	Name        string
	Version     string
	Description string
	Endpoints   []Endpoint
	BaseURL     string
	Timeout     int // in milliseconds
	RetryCount  int
	IsActive    bool
	Metadata    map[string]string
}

// Endpoint represents an endpoint of a backend service
type Endpoint struct {
	Path        string
	Methods     []string
	RateLimit   int // requests per minute
	AuthRequired bool
	Timeout     int // in milliseconds, overrides service timeout if set
	CacheTTL    int // in seconds, 0 means no caching
}

// NewService creates a new Service entity
func NewService(id, name, version, description, baseURL string) *Service {
	return &Service{
		ID:          id,
		Name:        name,
		Version:     version,
		Description: description,
		BaseURL:     baseURL,
		Timeout:     5000, // Default 5 seconds
		RetryCount:  3,    // Default 3 retries
		IsActive:    true,
		Endpoints:   []Endpoint{},
		Metadata:    make(map[string]string),
	}
}

// AddEndpoint adds an endpoint to the service
func (s *Service) AddEndpoint(endpoint Endpoint) {
	s.Endpoints = append(s.Endpoints, endpoint)
}

// SetActive sets the active status of the service
func (s *Service) SetActive(active bool) {
	s.IsActive = active
}

// SetTimeout sets the timeout for the service
func (s *Service) SetTimeout(timeout int) {
	s.Timeout = timeout
}

// SetRetryCount sets the retry count for the service
func (s *Service) SetRetryCount(retryCount int) {
	s.RetryCount = retryCount
}

// AddMetadata adds metadata to the service
func (s *Service) AddMetadata(key, value string) {
	s.Metadata[key] = value
}
```

### Repository Interfaces

Repository interfaces define how the domain layer interacts with data sources.

#### `service_repo.go`

```go
package repository

import (
	"context"

	"api-gateway/internal/domain/entity"
)

// ServiceRepository defines the interface for service repository
type ServiceRepository interface {
	// GetAll returns all registered services
	GetAll(ctx context.Context) ([]*entity.Service, error)
	
	// GetByID returns a service by its ID
	GetByID(ctx context.Context, id string) (*entity.Service, error)
	
	// GetByName returns a service by its name
	GetByName(ctx context.Context, name string) (*entity.Service, error)
	
	// GetByEndpoint returns services that match the given path and method
	GetByEndpoint(ctx context.Context, path, method string) ([]*entity.Service, error)
	
	// Create creates a new service
	Create(ctx context.Context, service *entity.Service) error
	
	// Update updates an existing service
	Update(ctx context.Context, service *entity.Service) error
	
	// Delete deletes a service by its ID
	Delete(ctx context.Context, id string) error
}
```

### Service Interfaces

Service interfaces define the core business operations of the API Gateway.

#### `gateway_service.go`

```go
package service

import (
	"context"

	"api-gateway/internal/domain/entity"
)

// GatewayService defines the interface for the API Gateway service
type GatewayService interface {
	// RouteRequest routes a request to the appropriate backend service
	RouteRequest(ctx context.Context, request *entity.Request) (*entity.Response, error)
	
	// ValidateRequest validates a request before routing
	ValidateRequest(ctx context.Context, request *entity.Request) error
	
	// TransformRequest transforms a request before sending to backend
	TransformRequest(ctx context.Context, request *entity.Request, service *entity.Service) (*entity.Request, error)
	
	// TransformResponse transforms a response before sending to client
	TransformResponse(ctx context.Context, response *entity.Response, service *entity.Service) (*entity.Response, error)
	
	// HandleError handles errors during request processing
	HandleError(ctx context.Context, err error, request *entity.Request) (*entity.Response, error)
}
```

#### `auth_service.go`

```go
package service

import (
	"context"

	"api-gateway/internal/domain/entity"
)

// AuthService defines the interface for authentication service
type AuthService interface {
	// Authenticate authenticates a request
	Authenticate(ctx context.Context, request *entity.Request) (bool, string, error)
	
	// Authorize authorizes a request for a specific service and endpoint
	Authorize(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) error
	
	// GenerateToken generates an authentication token
	GenerateToken(ctx context.Context, userID string, claims map[string]interface{}) (string, error)
	
	// ValidateToken validates an authentication token
	ValidateToken(ctx context.Context, token string) (map[string]interface{}, error)
}
```

#### `rate_limit_service.go`

```go
package service

import (
	"context"

	"api-gateway/internal/domain/entity"
)

// RateLimitService defines the interface for rate limiting service
type RateLimitService interface {
	// CheckLimit checks if a request exceeds the rate limit
	CheckLimit(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) (bool, error)
	
	// RecordRequest records a request for rate limiting purposes
	RecordRequest(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) error
	
	// GetLimit gets the current rate limit for a client
	GetLimit(ctx context.Context, clientID string, service *entity.Service, endpoint *entity.Endpoint) (int, int, error)
}
```

#### `cache_service.go`

```go
package service

import (
	"context"
	"time"

	"api-gateway/internal/domain/entity"
)

// CacheService defines the interface for caching service
type CacheService interface {
	// Get gets a cached response for a request
	Get(ctx context.Context, request *entity.Request) (*entity.Response, bool, error)
	
	// Set caches a response for a request
	Set(ctx context.Context, request *entity.Request, response *entity.Response, ttl time.Duration) error
	
	// Delete deletes a cached response for a request
	Delete(ctx context.Context, request *entity.Request) error
	
	// Clear clears all cached responses
	Clear(ctx context.Context) error
}
```

## Error Types

The domain layer defines custom error types for different failure scenarios.

```go
package errors

import (
	"fmt"
)

// ErrorType represents the type of an error
type ErrorType string

const (
	// ErrorTypeNotFound represents a not found error
	ErrorTypeNotFound ErrorType = "NOT_FOUND"
	
	// ErrorTypeValidation represents a validation error
	ErrorTypeValidation ErrorType = "VALIDATION"
	
	// ErrorTypeAuthentication represents an authentication error
	ErrorTypeAuthentication ErrorType = "AUTHENTICATION"
	
	// ErrorTypeAuthorization represents an authorization error
	ErrorTypeAuthorization ErrorType = "AUTHORIZATION"
	
	// ErrorTypeRateLimit represents a rate limit error
	ErrorTypeRateLimit ErrorType = "RATE_LIMIT"
	
	// ErrorTypeTimeout represents a timeout error
	ErrorTypeTimeout ErrorType = "TIMEOUT"
	
	// ErrorTypeCircuitBreaker represents a circuit breaker error
	ErrorTypeCircuitBreaker ErrorType = "CIRCUIT_BREAKER"
	
	// ErrorTypeInternal represents an internal error
	ErrorTypeInternal ErrorType = "INTERNAL"
)

// DomainError represents a domain error
type DomainError struct {
	Type    ErrorType
	Message string
	Cause   error
}

// Error returns the error message
func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %s)", e.Type, e.Message, e.Cause.Error())
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string, cause error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeNotFound,
		Message: message,
		Cause:   cause,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string, cause error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeValidation,
		Message: message,
		Cause:   cause,
	}
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(message string, cause error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeAuthentication,
		Message: message,
		Cause:   cause,
	}
}

// NewAuthorizationError creates a new authorization error
func NewAuthorizationError(message string, cause error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeAuthorization,
		Message: message,
		Cause:   cause,
	}
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(message string, cause error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeRateLimit,
		Message: message,
		Cause:   cause,
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(message string, cause error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeTimeout,
		Message: message,
		Cause:   cause,
	}
}

// NewCircuitBreakerError creates a new circuit breaker error
func NewCircuitBreakerError(message string, cause error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeCircuitBreaker,
		Message: message,
		Cause:   cause,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, cause error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeInternal,
		Message: message,
		Cause:   cause,
	}
}
```

## Domain Events

The domain layer defines events that can be published and subscribed to by other layers.

```go
package event

import (
	"time"
)

// Event represents a domain event
type Event interface {
	// EventType returns the type of the event
	EventType() string
	
	// Timestamp returns the timestamp of the event
	Timestamp() time.Time
	
	// Payload returns the payload of the event
	Payload() interface{}
}

// BaseEvent represents a base implementation of Event
type BaseEvent struct {
	Type      string
	Time      time.Time
	Data      interface{}
}

// EventType returns the type of the event
func (e *BaseEvent) EventType() string {
	return e.Type
}

// Timestamp returns the timestamp of the event
func (e *BaseEvent) Timestamp() time.Time {
	return e.Time
}

// Payload returns the payload of the event
func (e *BaseEvent) Payload() interface{} {
	return e.Data
}

// NewEvent creates a new event
func NewEvent(eventType string, payload interface{}) Event {
	return &BaseEvent{
		Type: eventType,
		Time: time.Now(),
		Data: payload,
	}
}

// RequestReceivedEvent represents a request received event
type RequestReceivedEvent struct {
	BaseEvent
}

// NewRequestReceivedEvent creates a new request received event
func NewRequestReceivedEvent(requestID string, path string, method string) Event {
	return &RequestReceivedEvent{
		BaseEvent: BaseEvent{
			Type: "REQUEST_RECEIVED",
			Time: time.Now(),
			Data: map[string]string{
				"requestID": requestID,
				"path":      path,
				"method":    method,
			},
		},
	}
}

// ResponseSentEvent represents a response sent event
type ResponseSentEvent struct {
	BaseEvent
}

// NewResponseSentEvent creates a new response sent event
func NewResponseSentEvent(requestID string, statusCode int, latencyMs int64) Event {
	return &ResponseSentEvent{
		BaseEvent: BaseEvent{
			Type: "RESPONSE_SENT",
			Time: time.Now(),
			Data: map[string]interface{}{
				"requestID":  requestID,
				"statusCode": statusCode,
				"latencyMs":  latencyMs,
			},
		},
	}
}

// ServiceDiscoveredEvent represents a service discovered event
type ServiceDiscoveredEvent struct {
	BaseEvent
}

// NewServiceDiscoveredEvent creates a new service discovered event
func NewServiceDiscoveredEvent(serviceID string, serviceName string) Event {
	return &ServiceDiscoveredEvent{
		BaseEvent: BaseEvent{
			Type: "SERVICE_DISCOVERED",
			Time: time.Now(),
			Data: map[string]string{
				"serviceID":   serviceID,
				"serviceName": serviceName,
			},
		},
	}
}

// ErrorOccurredEvent represents an error occurred event
type ErrorOccurredEvent struct {
	BaseEvent
}

// NewErrorOccurredEvent creates a new error occurred event
func NewErrorOccurredEvent(requestID string, errorType string, errorMessage string) Event {
	return &ErrorOccurredEvent{
		BaseEvent: BaseEvent{
			Type: "ERROR_OCCURRED",
			Time: time.Now(),
			Data: map[string]string{
				"requestID":    requestID,
				"errorType":    errorType,
				"errorMessage": errorMessage,
			},
		},
	}
}
```

## Value Objects

The domain layer defines value objects that are immutable and identified by their attributes.

```go
package valueobject

import (
	"net/url"
	"strings"
)

// Route represents a routing rule
type Route struct {
	Path       string
	PathRegex  string
	Methods    []string
	ServiceID  string
	Priority   int
	StripPath  bool
	AddHeaders map[string]string
}

// NewRoute creates a new Route
func NewRoute(path string, methods []string, serviceID string) *Route {
	return &Route{
		Path:       path,
		Methods:    methods,
		ServiceID:  serviceID,
		Priority:   0,
		StripPath:  false,
		AddHeaders: make(map[string]string),
	}
}

// MatchesPath checks if the route matches the given path
func (r *Route) MatchesPath(path string) bool {
	// Implementation details
	return false
}

// MatchesMethod checks if the route matches the given method
func (r *Route) MatchesMethod(method string) bool {
	for _, m := range r.Methods {
		if strings.EqualFold(m, method) || m == "*" {
			return true
		}
	}
	return false
}

// TransformPath transforms the path according to the route rules
func (r *Route) TransformPath(path string) string {
	if r.StripPath {
		// Strip the matching prefix
		return strings.TrimPrefix(path, r.Path)
	}
	return path
}

// APIKey represents an API key
type APIKey struct {
	Key       string
	ClientID  string
	ExpiresAt int64
	Scopes    []string
	RateLimit int
}

// NewAPIKey creates a new APIKey
func NewAPIKey(key, clientID string, expiresAt int64, scopes []string, rateLimit int) *APIKey {
	return &APIKey{
		Key:       key,
		ClientID:  clientID,
		ExpiresAt: expiresAt,
		Scopes:    scopes,
		RateLimit: rateLimit,
	}
}

// IsExpired checks if the API key is expired
func (k *APIKey) IsExpired(now int64) bool {
	return k.ExpiresAt > 0 && now > k.ExpiresAt
}

// HasScope checks if the API key has the given scope
func (k *APIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}

// Endpoint represents a URL endpoint
type Endpoint struct {
	URL         *url.URL
	Path        string
	QueryParams url.Values
}

// NewEndpoint creates a new Endpoint
func NewEndpoint(urlStr string) (*Endpoint, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	
	return &Endpoint{
		URL:         parsedURL,
		Path:        parsedURL.Path,
		QueryParams: parsedURL.Query(),
	}, nil
}

// AddQueryParam adds a query parameter to the endpoint
func (e *Endpoint) AddQueryParam(key, value string) {
	e.QueryParams.Add(key, value)
	e.URL.RawQuery = e.QueryParams.Encode()
}

// String returns the string representation of the endpoint
func (e *Endpoint) String() string {
	return e.URL.String()
}
```

## Domain Services

The domain layer defines domain services that implement core business logic.

```go
package domainservice

import (
	"context"
	"regexp"
	"strings"

	"api-gateway/internal/domain/entity"
	"api-gateway/internal/domain/valueobject"
)

// RouteMatcherService matches requests to routes
type RouteMatcherService struct {
	routes []*valueobject.Route
}

// NewRouteMatcherService creates a new RouteMatcherService
func NewRouteMatcherService() *RouteMatcherService {
	return &RouteMatcherService{
		routes: []*valueobject.Route{},
	}
}

// AddRoute adds a route to the matcher
func (s *RouteMatcherService) AddRoute(route *valueobject.Route) {
	s.routes = append(s.routes, route)
}

// MatchRoute matches a request to a route
func (s *RouteMatcherService) MatchRoute(ctx context.Context, request *entity.Request) (*valueobject.Route, bool) {
	// Sort routes by priority
	// Implementation details
	
	// Find the first matching route
	for _, route := range s.routes {
		if route.MatchesPath(request.Path) && route.MatchesMethod(request.Method) {
			return route, true
		}
	}
	
	return nil, false
}

// RequestValidatorService validates requests
type RequestValidatorService struct {
	maxBodySize      int64
	allowedMethods   []string
	forbiddenHeaders []string
	pathRegexes      map[string]*regexp.Regexp
}

// NewRequestValidatorService creates a new RequestValidatorService
func NewRequestValidatorService() *RequestValidatorService {
	return &RequestValidatorService{
		maxBodySize:      10 * 1024 * 1024, // 10MB
		allowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
		forbiddenHeaders: []string{"Host", "Connection", "Keep-Alive", "Transfer-Encoding", "TE", "Trailer", "Upgrade"},
		pathRegexes:      make(map[string]*regexp.Regexp),
	}
}

// ValidateRequest validates a request
func (s *RequestValidatorService) ValidateRequest(ctx context.Context, request *entity.Request) error {
	// Validate method
	methodAllowed := false
	for _, method := range s.allowedMethods {
		if strings.EqualFold(request.Method, method) {
			methodAllowed = true
			break
		}
	}
	if !methodAllowed {
		return NewValidationError("Method not allowed", nil)
	}
	
	// Validate body size
	if int64(len(request.Body)) > s.maxBodySize {
		return NewValidationError("Request body too large", nil)
	}
	
	// Validate headers
	for _, forbidden := range s.forbiddenHeaders {
		if _, exists := request.Headers[forbidden]; exists {
			delete(request.Headers, forbidden)
		}
	}
	
	// Validate path
	// Implementation details
	
	return nil
}

// NewValidationError creates a new validation error
func NewValidationError(message string, cause error) error {
	// Implementation details
	return nil
}
```

## Conclusion

The domain layer provides the core business entities and rules for the API Gateway. It defines the interfaces that outer layers must implement and contains no dependencies on other layers. This ensures that the business logic is isolated from infrastructure concerns and can be tested independently.
