# Main Application Entry Point

This document outlines the implementation details for the main application entry point of the API Gateway following Clean Architecture principles.

## Overview

The main application entry point ties together all layers of the API Gateway and provides the starting point for the application. It includes dependency injection, configuration loading, and server initialization.

## Main Application

### `main.go`

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"api-gateway/internal/application/usecase"
	"api-gateway/internal/domain/repository"
	"api-gateway/internal/domain/service"
	"api-gateway/internal/infrastructure/auth"
	"api-gateway/internal/infrastructure/cache"
	"api-gateway/internal/infrastructure/client"
	"api-gateway/internal/infrastructure/config"
	"api-gateway/internal/infrastructure/logger"
	"api-gateway/internal/infrastructure/metrics"
	"api-gateway/internal/infrastructure/persistence"
	infraRepo "api-gateway/internal/infrastructure/repository"
	"api-gateway/internal/infrastructure/ratelimit"
	"api-gateway/internal/interfaces/api"
	"api-gateway/pkg/errors"
)

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

	// Initialize schema
	err = persistence.InitSchema(db)
	if err != nil {
		appLogger.Error("Failed to initialize schema", "error", err)
		os.Exit(1)
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		appLogger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}

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

// newGatewayService creates a new gateway service
func newGatewayService(httpClient *client.HTTPClient, serviceRepo repository.ServiceRepository, logger logger.Logger) service.GatewayService {
	return &gatewayServiceImpl{
		httpClient:  httpClient,
		serviceRepo: serviceRepo,
		logger:      logger,
	}
}

// gatewayServiceImpl implements the GatewayService interface
type gatewayServiceImpl struct {
	httpClient  *client.HTTPClient
	serviceRepo repository.ServiceRepository
	logger      logger.Logger
}

// RouteRequest routes a request to the appropriate backend service
func (s *gatewayServiceImpl) RouteRequest(ctx context.Context, request *entity.Request) (*entity.Response, error) {
	// Find service for the request
	services, err := s.serviceRepo.GetByEndpoint(ctx, request.Path, request.Method)
	if err != nil {
		s.logger.Error("Failed to get service", "error", err)
		return nil, err
	}

	if len(services) == 0 {
		s.logger.Error("No service found for request", "path", request.Path, "method", request.Method)
		return nil, errors.NewNotFoundError("no service found for request", nil)
	}

	// Use the first active service
	var service *entity.Service
	for _, svc := range services {
		if svc.IsActive {
			service = svc
			break
		}
	}

	if service == nil {
		s.logger.Error("No active service found for request", "path", request.Path, "method", request.Method)
		return nil, errors.NewNotFoundError("no active service found for request", nil)
	}

	// Send request to backend service
	response, err := s.httpClient.SendRequest(ctx, request, service)
	if err != nil {
		s.logger.Error("Failed to send request to backend service", "error", err)
		return nil, err
	}

	return response, nil
}

// ValidateRequest validates a request before routing
func (s *gatewayServiceImpl) ValidateRequest(ctx context.Context, request *entity.Request) error {
	// Validate method
	if request.Method == "" {
		return errors.NewValidationError("method is required", nil)
	}

	// Validate path
	if request.Path == "" {
		return errors.NewValidationError("path is required", nil)
	}

	return nil
}

// TransformRequest transforms a request before sending to backend
func (s *gatewayServiceImpl) TransformRequest(ctx context.Context, request *entity.Request, service *entity.Service) (*entity.Request, error) {
	// Clone request
	transformedRequest := *request

	// Add headers
	if transformedRequest.Headers == nil {
		transformedRequest.Headers = make(map[string][]string)
	}

	// Add X-Forwarded-For header
	if clientIP := request.ClientIP; clientIP != "" {
		transformedRequest.Headers.Add("X-Forwarded-For", clientIP)
	}

	// Add X-Request-ID header if not present
	if requestID := request.ID; requestID != "" {
		transformedRequest.Headers.Set("X-Request-ID", requestID)
	} else {
		transformedRequest.Headers.Set("X-Request-ID", uuid.New().String())
	}

	// Add X-Forwarded-Host header
	if host := request.Headers.Get("Host"); host != "" {
		transformedRequest.Headers.Set("X-Forwarded-Host", host)
	}

	// Remove hop-by-hop headers
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"TE",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	for _, header := range hopByHopHeaders {
		transformedRequest.Headers.Del(header)
	}

	return &transformedRequest, nil
}

// TransformResponse transforms a response before sending to client
func (s *gatewayServiceImpl) TransformResponse(ctx context.Context, response *entity.Response, service *entity.Service) (*entity.Response, error) {
	// Clone response
	transformedResponse := *response

	// Add headers
	if transformedResponse.Headers == nil {
		transformedResponse.Headers = make(map[string][]string)
	}

	// Add X-Gateway header
	transformedResponse.Headers.Set("X-Gateway", "API-Gateway")

	// Remove hop-by-hop headers
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"TE",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	for _, header := range hopByHopHeaders {
		transformedResponse.Headers.Del(header)
	}

	return &transformedResponse, nil
}

// HandleError handles errors during request processing
func (s *gatewayServiceImpl) HandleError(ctx context.Context, err error, request *entity.Request) (*entity.Response, error) {
	// Create error response
	var statusCode int
	var errorMessage string

	// Check error type
	switch e := err.(type) {
	case *errors.DomainError:
		switch e.Type {
		case errors.ErrorTypeNotFound:
			statusCode = http.StatusNotFound
			errorMessage = "Not Found"
		case errors.ErrorTypeValidation:
			statusCode = http.StatusBadRequest
			errorMessage = "Bad Request"
		case errors.ErrorTypeAuthentication:
			statusCode = http.StatusUnauthorized
			errorMessage = "Unauthorized"
		case errors.ErrorTypeAuthorization:
			statusCode = http.StatusForbidden
			errorMessage = "Forbidden"
		case errors.ErrorTypeRateLimit:
			statusCode = http.StatusTooManyRequests
			errorMessage = "Too Many Requests"
		case errors.ErrorTypeTimeout:
			statusCode = http.StatusGatewayTimeout
			errorMessage = "Gateway Timeout"
		case errors.ErrorTypeCircuitBreaker:
			statusCode = http.StatusServiceUnavailable
			errorMessage = "Service Unavailable"
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Internal Server Error"
		}
	default:
		statusCode = http.StatusInternalServerError
		errorMessage = "Internal Server Error"
	}

	// Create error body
	errorBody := map[string]string{
		"error":   errorMessage,
		"message": err.Error(),
	}

	errorBodyBytes, _ := json.Marshal(errorBody)

	// Create response
	response := entity.NewResponse(
		request.ID,
		statusCode,
		map[string][]string{
			"Content-Type": {"application/json"},
		},
		errorBodyBytes,
	)

	return response, nil
}
```

### `go.mod`

```
module api-gateway

go 1.20

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-sql-driver/mysql v1.7.1
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.17
	github.com/prometheus/client_golang v1.16.0
	github.com/spf13/viper v1.16.0
	go.uber.org/zap v1.24.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
```

### Configuration File

#### `config.yaml`

```yaml
server:
  port: 8080
  read_timeout: 5s
  write_timeout: 10s
  shutdown_timeout: 30s

database:
  driver: sqlite3
  database: api_gateway.db

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

auth:
  secret_key: your-secret-key-here
  issuer: api-gateway
  expiration: 1h

logging:
  level: info
  development: true
```

### Docker Configuration

#### `Dockerfile`

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

#### `docker-compose.yml`

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

### Build and Run Scripts

#### `scripts/build.sh`

```bash
#!/bin/bash

# Build the application
echo "Building API Gateway..."
go build -o bin/api-gateway cmd/api/main.go

if [ $? -eq 0 ]; then
    echo "Build successful!"
else
    echo "Build failed!"
    exit 1
fi
```

#### `scripts/run.sh`

```bash
#!/bin/bash

# Run the application
echo "Running API Gateway..."
./bin/api-gateway
```

## Dependency Injection

The main application entry point uses manual dependency injection to wire up all the components of the API Gateway. This approach provides several benefits:

1. **Explicit Dependencies**: Dependencies are explicitly declared, making it clear what each component needs.
2. **Testability**: Components can be easily mocked for testing.
3. **Flexibility**: Components can be easily replaced with different implementations.

The dependency injection flow is as follows:

1. **Configuration**: Load configuration from file and environment variables.
2. **Logger**: Initialize the logger.
3. **Metrics**: Initialize the metrics collector.
4. **Database**: Initialize the database connection.
5. **Redis**: Initialize the Redis connection.
6. **Cache**: Initialize the cache service.
7. **Repositories**: Initialize the repositories.
8. **HTTP Client**: Initialize the HTTP client.
9. **Authentication**: Initialize the authentication service.
10. **Rate Limiting**: Initialize the rate limiting service.
11. **Gateway Service**: Initialize the gateway service.
12. **Use Cases**: Initialize the use cases.
13. **Handler**: Initialize the handler.
14. **Router**: Initialize the router.
15. **Server**: Initialize the server.

## Configuration Management

The API Gateway uses a combination of configuration files and environment variables for configuration management. The configuration is loaded in the following order:

1. **Default Values**: Default values are set in the code.
2. **Configuration File**: Values from the configuration file override default values.
3. **Environment Variables**: Values from environment variables override values from the configuration file.

This approach provides flexibility in how the application is configured, allowing for different configurations in different environments.

## Error Handling

The API Gateway implements robust error handling to ensure that errors are properly handled and reported. The error handling flow is as follows:

1. **Error Detection**: Errors are detected at the point where they occur.
2. **Error Logging**: Errors are logged with relevant context information.
3. **Error Transformation**: Errors are transformed into appropriate HTTP responses.
4. **Error Reporting**: Errors are reported to the client with appropriate status codes and messages.

## Conclusion

The main application entry point ties together all layers of the API Gateway and provides the starting point for the application. It includes dependency injection, configuration loading, and server initialization.

Key components of the main application entry point include:

1. **Main Function**: The entry point of the application.
2. **Dependency Injection**: Wires up all the components of the API Gateway.
3. **Configuration Management**: Loads configuration from file and environment variables.
4. **Error Handling**: Ensures that errors are properly handled and reported.

The main application entry point follows clean architecture principles, with clear separation of concerns and dependencies flowing inward.
