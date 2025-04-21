# API Gateway Implementation Plan

## Table of Contents

1. [Introduction](#introduction)
2. [Clean Architecture Overview](#clean-architecture-overview)
3. [Project Structure](#project-structure)
4. [Architecture Design](#architecture-design)
5. [Domain Layer](#domain-layer)
6. [Application Layer](#application-layer)
7. [Infrastructure Layer](#infrastructure-layer)
8. [Interfaces Layer](#interfaces-layer)
9. [Main Application](#main-application)
10. [Deployment](#deployment)
11. [Testing](#testing)
12. [Conclusion](#conclusion)

## Introduction

This document outlines a comprehensive implementation plan for an API Gateway using Golang and following Clean Architecture principles. An API Gateway serves as a single entry point for client applications to access various backend services, providing cross-cutting concerns like routing, authentication, rate limiting, and more.

The implementation follows Clean Architecture to ensure separation of concerns, maintainability, and testability. This approach allows for independent development and testing of different components, making the system more robust and adaptable to change.

## Clean Architecture Overview

Clean Architecture, proposed by Robert C. Martin, emphasizes separation of concerns through layers:

1. **Domain Layer** - Contains business logic and entities
2. **Application Layer** - Contains use cases and application-specific business rules
3. **Infrastructure Layer** - Contains implementations of interfaces defined in inner layers
4. **Interfaces Layer** - Contains controllers, presenters, and gateways

The dependencies flow inward, with inner layers having no knowledge of outer layers. This ensures that business logic is isolated from external concerns like databases, frameworks, and UI.

![Clean Architecture Diagram](https://blog.cleancoder.com/uncle-bob/images/2012-08-13-the-clean-architecture/CleanArchitecture.jpg)

## Project Structure

The project structure follows Clean Architecture principles, with clear separation between layers:

```
api-gateway/
├── cmd/                          # Application entry points
│   └── api/                      # API Gateway service
│       └── main.go               # Main application entry point
├── internal/                     # Private application code
│   ├── domain/                   # Domain layer
│   │   ├── entity/               # Business entities
│   │   │   └── request.go        # Request entity
│   │   │   └── response.go       # Response entity
│   │   │   └── service.go        # Service entity
│   │   ├── repository/           # Repository interfaces
│   │   │   └── service_repo.go   # Service repository interface
│   │   └── service/              # Domain service interfaces
│   │       └── gateway_service.go # Gateway service interface
│   ├── application/              # Application layer
│   │   ├── usecase/              # Use cases
│   │   │   └── proxy_usecase.go  # Proxy request use case
│   │   │   └── auth_usecase.go   # Authentication use case
│   │   │   └── rate_limit_usecase.go # Rate limiting use case
│   │   └── dto/                  # Data Transfer Objects
│   │       └── request_dto.go    # Request DTO
│   │       └── response_dto.go   # Response DTO
│   ├── infrastructure/           # Infrastructure layer
│   │   ├── repository/           # Repository implementations
│   │   │   └── service_repo_impl.go # Service repository implementation
│   │   ├── client/               # External service clients
│   │   │   └── http_client.go    # HTTP client
│   │   ├── cache/                # Cache implementations
│   │   │   └── redis_cache.go    # Redis cache implementation
│   │   ├── auth/                 # Authentication implementations
│   │   │   └── jwt_auth.go       # JWT authentication
│   │   └── persistence/          # Database connections
│   │       └── database.go       # Database connection
│   └── interfaces/               # Interfaces layer
│       ├── api/                  # API handlers
│       │   └── handler.go        # API handlers
│       │   └── middleware.go     # API middlewares
│       │   └── router.go         # API router
│       └── dto/                  # Interface DTOs
│           └── api_dto.go        # API DTOs
├── pkg/                          # Public libraries
│   ├── logger/                   # Logging utilities
│   │   └── logger.go             # Logger implementation
│   ├── config/                   # Configuration utilities
│   │   └── config.go             # Configuration loader
│   └── errors/                   # Error handling utilities
│       └── errors.go             # Custom error types
├── configs/                      # Configuration files
│   └── config.yaml               # Application configuration
├── scripts/                      # Build and deployment scripts
│   └── build.sh                  # Build script
├── test/                         # Test files
│   ├── unit/                     # Unit tests
│   └── integration/              # Integration tests
├── docs/                         # Documentation
│   └── api.md                    # API documentation
├── go.mod                        # Go modules file
└── go.sum                        # Go modules checksum file
```

## Architecture Design

The API Gateway architecture consists of several core components that work together to provide a comprehensive solution:

### Request Processing Pipeline

The API Gateway processes requests through a pipeline of middleware components:

```
Client Request → Router → Authentication → Rate Limiting → Request Transformation → Service Discovery → Load Balancing → Backend Service → Response Transformation → Client Response
```

### Key Components

#### Router
- Responsible for routing incoming requests to appropriate handlers
- Implements path-based and method-based routing
- Supports versioning through URL paths or headers

#### Authentication & Authorization
- Validates client credentials (API keys, JWT tokens)
- Implements role-based access control
- Supports multiple authentication methods (Basic, OAuth2, JWT)

#### Rate Limiting
- Protects backend services from excessive requests
- Implements token bucket or leaky bucket algorithms
- Configurable per client, endpoint, or service

#### Service Discovery
- Maintains registry of available backend services
- Supports static configuration and dynamic discovery
- Integrates with service discovery tools (Consul, etcd)

#### Load Balancing
- Distributes requests across multiple instances of backend services
- Implements various strategies (Round Robin, Least Connections, Weighted)
- Handles health checks and circuit breaking

#### Request/Response Transformation
- Transforms client requests to backend service format
- Transforms backend responses to client format
- Handles protocol translation (REST to gRPC, etc.)

#### Caching
- Caches responses to reduce backend load
- Implements TTL-based invalidation
- Supports distributed caching with Redis

#### Logging & Monitoring
- Logs request/response details for auditing
- Collects metrics for monitoring
- Integrates with observability tools (Prometheus, Grafana)

#### Circuit Breaker
- Prevents cascading failures
- Implements fallback mechanisms
- Supports automatic recovery

### Architecture Diagram

```
┌─────────────┐     ┌─────────────────────────────────────────────────────────────────────┐     ┌─────────────┐
│             │     │                           API Gateway                                │     │             │
│             │     │  ┌─────────┐  ┌─────┐  ┌─────────┐  ┌─────────┐  ┌───────────────┐  │     │             │
│   Clients   │────▶│  │ Router  │─▶│Auth │─▶│  Rate   │─▶│ Service │─▶│ Load Balancer │──│────▶│  Backend   │
│  (Web/App)  │     │  │         │  │     │  │ Limiter │  │Discovery│  │               │  │     │  Services  │
│             │◀────│──│         │◀─│     │◀─│         │◀─│         │◀─│               │◀─│─────│             │
└─────────────┘     │  └─────────┘  └─────┘  └─────────┘  └─────────┘  └───────────────┘  │     └─────────────┘
                    │       ▲           ▲          ▲            ▲              ▲          │
                    │       │           │          │            │              │          │
                    │       │           │          │            │              │          │
                    │  ┌────┴───────────┴──────────┴────────────┴──────────────┴─────┐   │
                    │  │                                                              │   │
                    │  │                      Shared Components                       │   │
                    │  │                                                              │   │
                    │  │   ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐  │   │
                    │  │   │ Logging │    │  Cache  │    │  Config │    │ Circuit │  │   │
                    │  │   │         │    │         │    │ Manager │    │ Breaker │  │   │
                    │  │   └─────────┘    └─────────┘    └─────────┘    └─────────┘  │   │
                    │  │                                                              │   │
                    │  └──────────────────────────────────────────────────────────────┘   │
                    │                                                                     │
                    └─────────────────────────────────────────────────────────────────────┘
```

## Domain Layer

The domain layer contains the core business entities and business rules of the API Gateway. It has no dependencies on other layers and defines interfaces that outer layers must implement.

### Core Entities

#### Request Entity

```go
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
```

#### Response Entity

```go
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
```

#### Service Entity

```go
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
	Path         string
	Methods      []string
	RateLimit    int  // requests per minute
	AuthRequired bool
	Timeout      int  // in milliseconds, overrides service timeout if set
	CacheTTL     int  // in seconds, 0 means no caching
}
```

### Repository Interfaces

```go
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

```go
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

// RateLimitService defines the interface for rate limiting service
type RateLimitService interface {
	// CheckLimit checks if a request exceeds the rate limit
	CheckLimit(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) (bool, error)
	
	// RecordRequest records a request for rate limiting purposes
	RecordRequest(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) error
	
	// GetLimit gets the current rate limit for a client
	GetLimit(ctx context.Context, clientID string, service *entity.Service, endpoint *entity.Endpoint) (int, int, error)
}

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

## Application Layer

The application layer contains use cases that orchestrate the flow of data to and from the domain entities and apply application-specific business rules.

### Data Transfer Objects (DTOs)

```go
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
```

### Use Cases

#### Proxy Use Case

```go
// ProxyUseCase implements the use case for proxying requests to backend services
type ProxyUseCase struct {
	serviceRepo      repository.ServiceRepository
	gatewayService   service.GatewayService
	authService      service.AuthService
	rateLimitService service.RateLimitService
	cacheService     service.CacheService
	logger           Logger
}

// ProxyRequest proxies a request to a backend service
func (uc *ProxyUseCase) ProxyRequest(ctx context.Context, requestDTO *dto.RequestDTO) (*dto.ResponseDTO, error) {
	// Convert DTO to domain entity
	request := requestDTO.ToEntity()
	
	// Validate request
	if err := uc.gatewayService.ValidateRequest(ctx, request); err != nil {
		return uc.handleError(ctx, err, request)
	}
	
	// Find service for the request
	service, endpoint, err := uc.findService(ctx, request)
	if err != nil {
		return uc.handleError(ctx, err, request)
	}
	
	// Check if authentication is required
	if endpoint.AuthRequired {
		authenticated, userID, err := uc.authService.Authenticate(ctx, request)
		if err != nil {
			return uc.handleError(ctx, err, request)
		}
		
		if !authenticated {
			return uc.handleError(ctx, fmt.Errorf("unauthorized"), request)
		}
		
		request.SetAuthenticated(authenticated, userID)
		
		// Authorize the request
		if err := uc.authService.Authorize(ctx, request, service, endpoint); err != nil {
			return uc.handleError(ctx, err, request)
		}
	}
	
	// Check rate limit
	if endpoint.RateLimit > 0 {
		allowed, err := uc.rateLimitService.CheckLimit(ctx, request, service, endpoint)
		if err != nil {
			return uc.handleError(ctx, err, request)
		}
		
		if !allowed {
			return uc.handleError(ctx, fmt.Errorf("rate limit exceeded"), request)
		}
		
		// Record the request for rate limiting
		if err := uc.rateLimitService.RecordRequest(ctx, request, service, endpoint); err != nil {
			// Continue processing even if recording fails
		}
	}
	
	// Check cache
	if endpoint.CacheTTL > 0 {
		cachedResponse, found, err := uc.cacheService.Get(ctx, request)
		if err == nil && found && cachedResponse != nil {
			return dto.FromEntity(cachedResponse), nil
		}
	}
	
	// Transform request
	transformedRequest, err := uc.gatewayService.TransformRequest(ctx, request, service)
	if err != nil {
		return uc.handleError(ctx, err, request)
	}
	
	// Route request to backend service
	response, err := uc.gatewayService.RouteRequest(ctx, transformedRequest)
	if err != nil {
		return uc.handleError(ctx, err, request)
	}
	
	// Transform response
	transformedResponse, err := uc.gatewayService.TransformResponse(ctx, response, service)
	if err != nil {
		return uc.handleError(ctx, err, request)
	}
	
	// Cache response if needed
	if endpoint.CacheTTL > 0 {
		if err := uc.cacheService.Set(ctx, request, transformedResponse, time.Duration(endpoint.CacheTTL)*time.Second); err != nil {
			// Continue processing even if cache storage fails
		}
	}
	
	// Return response DTO
	return dto.FromEntity(transformedResponse), nil
}
```

## Infrastructure Layer

The infrastructure layer provides concrete implementations of the interfaces defined in the domain layer.

### Repository Implementations

```go
// ServiceRepositoryImpl implements the ServiceRepository interface
type ServiceRepositoryImpl struct {
	db     *sql.DB
	cache  Cache
	logger Logger
}

// GetAll returns all registered services
func (r *ServiceRepositoryImpl) GetAll(ctx context.Context) ([]*entity.Service, error) {
	// Try to get from cache
	if r.cache != nil {
		cachedServices, found := r.cache.Get(ctx, "services:all")
		if found {
			return cachedServices.([]*entity.Service), nil
		}
	}

	// Query database
	// Implementation details...

	// Cache the result
	if r.cache != nil {
		r.cache.Set(ctx, "services:all", services, 5*time.Minute)
	}

	return services, nil
}
```

### External Service Clients

```go
// HTTPClient implements an HTTP client for communicating with backend services
type HTTPClient struct {
	client  *http.Client
	logger  Logger
	metrics MetricsRecorder
}

// SendRequest sends an HTTP request to a backend service
func (c *HTTPClient) SendRequest(ctx context.Context, request *entity.Request, service *entity.Service) (*entity.Response, error) {
	startTime := time.Now()

	// Create target URL
	targetURL := fmt.Sprintf("%s%s", service.BaseURL, request.Path)
	
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, request.Method, targetURL, bytes.NewReader(request.Body))
	if err != nil {
		return nil, err
	}

	// Set headers, query parameters, etc.
	// Implementation details...

	// Send request
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	// Create response
	response := entity.NewResponse(
		request.ID,
		httpResp.StatusCode,
		httpResp.Header,
		body,
	)

	// Set latency
	response.SetLatency(startTime)

	// Record metrics
	if c.metrics != nil {
		c.metrics.RecordServiceLatency(service.ID, response.LatencyMs)
	}

	return response, nil
}
```

### Cache Implementations

```go
// RedisCache implements the CacheService interface using Redis
type RedisCache struct {
	client *redis.Client
	logger Logger
}

// Get gets a cached response for a request
func (c *RedisCache) Get(ctx context.Context, request *entity.Request) (*entity.Response, bool, error) {
	// Generate cache key
	key := c.generateCacheKey(request)

	// Get from Redis
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			// Key not found
			return nil, false, nil
		}
		return nil, false, err
	}

	// Unmarshal response
	var response entity.Response
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, false, err
	}

	// Mark as cached
	response.SetCached(true)

	return &response, true, nil
}
```

### Authentication Implementations

```go
// JWTAuth implements the AuthService interface using JWT
type JWTAuth struct {
	secretKey     []byte
	issuer        string
	expiration    time.Duration
	logger        Logger
	apiKeyRepo    APIKeyRepository
}

// Authenticate authenticates a request
func (a *JWTAuth) Authenticate(ctx context.Context, request *entity.Request) (bool, string, error) {
	// Check for API key in header
	apiKeyHeader := request.Headers.Get("X-API-Key")
	if apiKeyHeader != "" {
		return a.authenticateWithAPIKey(ctx, apiKeyHeader)
	}

	// Check for JWT token in Authorization header
	authHeader := request.Headers.Get("Authorization")
	if authHeader != "" {
		return a.authenticateWithJWT(ctx, authHeader)
	}

	return false, "", nil
}
```

## Interfaces Layer

The interfaces layer handles the communication between the outside world and the application.

### API Handlers

```go
// Handler handles HTTP requests
type Handler struct {
	proxyUseCase            *usecase.ProxyUseCase
	authUseCase             *usecase.AuthUseCase
	rateLimitUseCase        *usecase.RateLimitUseCase
	serviceManagementUseCase *usecase.ServiceManagementUseCase
	logger                  Logger
}

// ProxyHandler handles proxy requests
func (h *Handler) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Create request DTO
	requestDTO := &dto.RequestDTO{
		Method:      r.Method,
		Path:        r.URL.Path,
		Headers:     r.Header,
		QueryParams: r.URL.Query(),
		Body:        body,
		ClientIP:    getClientIP(r),
		Timestamp:   time.Now(),
	}

	// Proxy request
	responseDTO, err := h.proxyUseCase.ProxyRequest(r.Context(), requestDTO)
	if err != nil {
		http.Error(w, "Failed to proxy request", http.StatusInternalServerError)
		return
	}

	// Set response headers
	for key, values := range responseDTO.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code
	w.WriteHeader(responseDTO.StatusCode)

	// Write response body
	w.Write(responseDTO.Body)
}
```

### Middleware

```go
// LoggingMiddleware logs requests
func LoggingMiddleware(logger Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Create a response writer wrapper to capture status code
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Generate request ID if not present
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
				r.Header.Set("X-Request-ID", requestID)
			}

			// Log request
			logger.Info("Request started",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			// Call next handler
			next.ServeHTTP(rw, r)

			// Log response
			logger.Info("Request completed",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"duration_ms", time.Since(startTime).Milliseconds(),
			)
		})
	}
}
```

### Router

```go
// Setup sets up the router
func (r *Router) Setup() http.Handler {
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(
		LoggingMiddleware(r.logger),
		RecoveryMiddleware(r.logger),
		CORSMiddleware([]string{"*"}),
		TimeoutMiddleware(30 * time.Second),
		MetricsMiddleware(r.metrics),
	)

	// API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(
		AuthMiddleware(r.authUseCase, r.logger),
		RateLimitMiddleware(r.rateLimitUseCase, r.logger),
	)

	// Proxy routes
	apiRouter.PathPrefix("/v1/").Handler(http.HandlerFunc(r.handler.ProxyHandler))

	// Auth routes
	router.HandleFunc("/auth", r.handler.AuthHandler).Methods(http.MethodPost)
	router.HandleFunc("/auth/validate", r.handler.ValidateTokenHandler).Methods(http.MethodGet)

	// Service management routes
	apiRouter.HandleFunc("/services", r.handler.GetServicesHandler).Methods(http.MethodGet)
	apiRouter.HandleFunc("/services/{id}", r.handler.GetServiceHandler).Methods(http.MethodGet)
	apiRouter.HandleFunc("/services", r.handler.CreateServiceHandler).Methods(http.MethodPost)
	apiRouter.HandleFunc("/services/{id}", r.handler.UpdateServiceHandler).Methods(http.MethodPut)
	apiRouter.HandleFunc("/services/{id}", r.handler.DeleteServiceHandler).Methods(http.MethodDelete)

	// Rate limit routes
	apiRouter.HandleFunc("/rate-limits", r.handler.CheckRateLimitHandler).Methods(http.MethodGet)

	// Health check route
	router.HandleFunc("/health", r.handler.HealthCheckHandler).Methods(http.MethodGet)

	return router
}
```

## Main Application

The main application entry point ties together all layers of the API Gateway and provides the starting point for the application.

```go
func main() {
	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.NewZapLogger(cfg.Logging.Level, cfg.Logging.Development)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	appLogger.Info("Starting API Gateway")

	// Initialize metrics
	prometheusMetrics := metrics.NewPrometheusMetrics()

	// Initialize database
	db, err := persistence.NewDatabase(cfg.Database)
	if err != nil {
		appLogger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Initialize cache
	cacheService := cache.NewRedisCache(redisClient, appLogger)

	// Initialize repositories
	serviceRepo := infraRepo.NewServiceRepositoryImpl(db, cacheService, appLogger)

	// Initialize HTTP client
	httpClient := client.NewHTTPClient(30*time.Second, appLogger, prometheusMetrics)

	// Initialize authentication service
	authService := auth.NewJWTAuth(
		[]byte(cfg.Auth.SecretKey),
		cfg.Auth.Issuer,
		cfg.Auth.Expiration,
		appLogger,
		nil, // API key repository not implemented yet
	)

	// Initialize rate limiting service
	rateLimitService := ratelimit.NewTokenBucketRateLimiter(redisClient, appLogger)

	// Initialize gateway service
	gatewayService := newGatewayService(httpClient, serviceRepo, appLogger)

	// Initialize use cases
	proxyUseCase := usecase.NewProxyUseCase(
		serviceRepo,
		gatewayService,
		authService,
		rateLimitService,
		cacheService,
		appLogger,
	)

	authUseCase := usecase.NewAuthUseCase(authService, appLogger)
	rateLimitUseCase := usecase.NewRateLimitUseCase(rateLimitService, appLogger)
	serviceManagementUseCase := usecase.NewServiceManagementUseCase(serviceRepo, appLogger)

	// Initialize handler
	handler := api.NewHandler(
		proxyUseCase,
		authUseCase,
		rateLimitUseCase,
		serviceManagementUseCase,
		appLogger,
	)

	// Initialize router
	router := api.NewRouter(
		handler,
		appLogger,
		prometheusMetrics,
		authUseCase,
		rateLimitUseCase,
	)

	// Initialize server
	server := api.NewServer(
		router.Setup(),
		cfg.Server.Port,
		cfg.Server.ReadTimeout,
		cfg.Server.WriteTimeout,
		cfg.Server.ShutdownTimeout,
		appLogger,
	)

	// Start server
	appLogger.Info("Server initialized", "port", cfg.Server.Port)
	if err := server.Start(); err != nil {
		appLogger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
```

## Deployment

The API Gateway can be deployed using Docker and Docker Compose for easy setup and scaling.

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway ./cmd/api/main.go

# Final stage
FROM alpine:3.18

WORKDIR /app

# Install dependencies
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder stage
COPY --from=builder /app/api-gateway .
COPY --from=builder /app/configs/config.yaml ./configs/

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api-gateway"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  api-gateway:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - API_GATEWAY_DATABASE_DRIVER=postgres
      - API_GATEWAY_DATABASE_HOST=postgres
      - API_GATEWAY_DATABASE_PORT=5432
      - API_GATEWAY_DATABASE_USER=postgres
      - API_GATEWAY_DATABASE_PASSWORD=postgres
      - API_GATEWAY_DATABASE_DATABASE=api_gateway
      - API_GATEWAY_REDIS_HOST=redis
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=api_gateway
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

## Testing

Testing is an essential part of the API Gateway implementation. The clean architecture approach makes it easier to test each layer independently.

### Unit Testing

Unit tests focus on testing individual components in isolation. For example, testing a use case with mocked dependencies:

```go
func TestProxyUseCase_ProxyRequest(t *testing.T) {
	// Create mocks
	mockServiceRepo := mocks.NewMockServiceRepository()
	mockGatewayService := mocks.NewMockGatewayService()
	mockAuthService := mocks.NewMockAuthService()
	mockRateLimitService := mocks.NewMockRateLimitService()
	mockCacheService := mocks.NewMockCacheService()
	mockLogger := mocks.NewMockLogger()

	// Create use case
	useCase := usecase.NewProxyUseCase(
		mockServiceRepo,
		mockGatewayService,
		mockAuthService,
		mockRateLimitService,
		mockCacheService,
		mockLogger,
	)

	// Create request DTO
	requestDTO := &dto.RequestDTO{
		Method: "GET",
		Path:   "/api/test",
		// Other fields...
	}

	// Set up expectations
	mockGatewayService.On("ValidateRequest", mock.Anything, mock.Anything).Return(nil)
	mockServiceRepo.On("GetByEndpoint", mock.Anything, "/api/test", "GET").Return([]*entity.Service{
		{
			ID:      "test-service",
			Name:    "Test Service",
			BaseURL: "http://test-service",
			Endpoints: []entity.Endpoint{
				{
					Path:         "/api/test",
					Methods:      []string{"GET"},
					AuthRequired: false,
				},
			},
		},
	}, nil)
	mockGatewayService.On("TransformRequest", mock.Anything, mock.Anything, mock.Anything).Return(&entity.Request{}, nil)
	mockGatewayService.On("RouteRequest", mock.Anything, mock.Anything).Return(&entity.Response{
		StatusCode: 200,
		Body:       []byte(`{"message":"success"}`),
	}, nil)
	mockGatewayService.On("TransformResponse", mock.Anything, mock.Anything, mock.Anything).Return(&entity.Response{
		StatusCode: 200,
		Body:       []byte(`{"message":"success"}`),
	}, nil)

	// Call the method
	responseDTO, err := useCase.ProxyRequest(context.Background(), requestDTO)

	// Assert expectations
	assert.NoError(t, err)
	assert.NotNil(t, responseDTO)
	assert.Equal(t, 200, responseDTO.StatusCode)
	assert.Equal(t, []byte(`{"message":"success"}`), responseDTO.Body)
	mockGatewayService.AssertExpectations(t)
	mockServiceRepo.AssertExpectations(t)
}
```

### Integration Testing

Integration tests focus on testing the interaction between components. For example, testing the API handlers with a test server:

```go
func TestHandler_ProxyHandler(t *testing.T) {
	// Create test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"success"}`))
	}))
	defer testServer.Close()

	// Create mocks
	mockServiceRepo := mocks.NewMockServiceRepository()
	mockGatewayService := mocks.NewMockGatewayService()
	mockAuthService := mocks.NewMockAuthService()
	mockRateLimitService := mocks.NewMockRateLimitService()
	mockCacheService := mocks.NewMockCacheService()
	mockLogger := mocks.NewMockLogger()

	// Create use case
	proxyUseCase := usecase.NewProxyUseCase(
		mockServiceRepo,
		mockGatewayService,
		mockAuthService,
		mockRateLimitService,
		mockCacheService,
		mockLogger,
	)

	// Create handler
	handler := api.NewHandler(
		proxyUseCase,
		nil, // Auth use case not needed for this test
		nil, // Rate limit use case not needed for this test
		nil, // Service management use case not needed for this test
		mockLogger,
	)

	// Create test request
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Content-Type", "application/json")

	// Create test response recorder
	rr := httptest.NewRecorder()

	// Set up expectations
	mockServiceRepo.On("GetByEndpoint", mock.Anything, "/api/test", "GET").Return([]*entity.Service{
		{
			ID:      "test-service",
			Name:    "Test Service",
			BaseURL: testServer.URL,
			Endpoints: []entity.Endpoint{
				{
					Path:         "/api/test",
					Methods:      []string{"GET"},
					AuthRequired: false,
				},
			},
		},
	}, nil)
	mockGatewayService.On("ValidateRequest", mock.Anything, mock.Anything).Return(nil)
	mockGatewayService.On("TransformRequest", mock.Anything, mock.Anything, mock.Anything).Return(&entity.Request{}, nil)
	mockGatewayService.On("RouteRequest", mock.Anything, mock.Anything).Return(&entity.Response{
		StatusCode: 200,
		Body:       []byte(`{"message":"success"}`),
	}, nil)
	mockGatewayService.On("TransformResponse", mock.Anything, mock.Anything, mock.Anything).Return(&entity.Response{
		StatusCode: 200,
		Body:       []byte(`{"message":"success"}`),
	}, nil)

	// Call the handler
	handler.ProxyHandler(rr, req)

	// Assert expectations
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message":"success"}`, rr.Body.String())
	mockGatewayService.AssertExpectations(t)
	mockServiceRepo.AssertExpectations(t)
}
```

## Conclusion

This implementation plan provides a comprehensive approach to building an API Gateway using Golang and following Clean Architecture principles. The key benefits of this approach include:

1. **Separation of Concerns**: Each layer has a specific responsibility, making the code more maintainable and easier to understand.
2. **Testability**: The clean architecture approach makes it easier to test each layer independently.
3. **Flexibility**: The use of interfaces allows for easy swapping of implementations, such as changing the database or cache provider.
4. **Scalability**: The modular design allows for easy scaling of individual components.

The API Gateway provides essential features for managing API traffic, including:

- Routing and proxying requests to backend services
- Authentication and authorization
- Rate limiting
- Service discovery
- Load balancing
- Request/response transformation
- Caching
- Logging and monitoring
- Circuit breaking

By following this implementation plan, you can build a robust and scalable API Gateway that meets the needs of modern microservice architectures.
