# Infrastructure Layer Implementation

This document outlines the implementation details for the infrastructure layer of the API Gateway following Clean Architecture principles.

## Overview

The infrastructure layer is the outermost layer of the Clean Architecture and provides concrete implementations of the interfaces defined in the domain layer. It includes implementations for repositories, external service clients, authentication, caching, and other infrastructure concerns.

## Core Components

### Repository Implementations

Repository implementations provide concrete implementations of the repository interfaces defined in the domain layer.

#### `service_repo_impl.go`

```go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"api-gateway/internal/domain/entity"
	domainRepo "api-gateway/internal/domain/repository"
)

// ServiceRepositoryImpl implements the ServiceRepository interface
type ServiceRepositoryImpl struct {
	db     *sql.DB
	cache  Cache
	logger Logger
}

// Cache defines the interface for caching
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewServiceRepositoryImpl creates a new ServiceRepositoryImpl
func NewServiceRepositoryImpl(db *sql.DB, cache Cache, logger Logger) domainRepo.ServiceRepository {
	return &ServiceRepositoryImpl{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// GetAll returns all registered services
func (r *ServiceRepositoryImpl) GetAll(ctx context.Context) ([]*entity.Service, error) {
	// Try to get from cache
	if r.cache != nil {
		cachedServices, found := r.cache.Get(ctx, "services:all")
		if found {
			r.logger.Debug("Cache hit for all services")
			return cachedServices.([]*entity.Service), nil
		}
	}

	// Query database
	query := `
		SELECT 
			s.id, s.name, s.version, s.description, s.base_url, s.timeout, s.retry_count, s.is_active,
			e.path, e.methods, e.rate_limit, e.auth_required, e.timeout, e.cache_ttl,
			m.key, m.value
		FROM services s
		LEFT JOIN endpoints e ON s.id = e.service_id
		LEFT JOIN service_metadata m ON s.id = m.service_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to query services", "error", err)
		return nil, err
	}
	defer rows.Close()

	// Map to store services by ID
	servicesMap := make(map[string]*entity.Service)

	// Iterate through rows
	for rows.Next() {
		var serviceID, serviceName, serviceVersion, serviceDescription, serviceBaseURL string
		var serviceTimeout, serviceRetryCount int
		var serviceIsActive bool
		var endpointPath, endpointMethods, metadataKey, metadataValue sql.NullString
		var endpointRateLimit, endpointTimeout, endpointCacheTTL sql.NullInt64
		var endpointAuthRequired sql.NullBool

		err := rows.Scan(
			&serviceID, &serviceName, &serviceVersion, &serviceDescription, &serviceBaseURL,
			&serviceTimeout, &serviceRetryCount, &serviceIsActive,
			&endpointPath, &endpointMethods, &endpointRateLimit, &endpointAuthRequired, &endpointTimeout, &endpointCacheTTL,
			&metadataKey, &metadataValue,
		)
		if err != nil {
			r.logger.Error("Failed to scan service row", "error", err)
			return nil, err
		}

		// Get or create service
		service, exists := servicesMap[serviceID]
		if !exists {
			service = entity.NewService(
				serviceID,
				serviceName,
				serviceVersion,
				serviceDescription,
				serviceBaseURL,
			)
			service.Timeout = serviceTimeout
			service.RetryCount = serviceRetryCount
			service.IsActive = serviceIsActive
			servicesMap[serviceID] = service
		}

		// Add endpoint if it exists
		if endpointPath.Valid {
			methods := []string{}
			if endpointMethods.Valid {
				methods = strings.Split(endpointMethods.String, ",")
			}

			rateLimit := 0
			if endpointRateLimit.Valid {
				rateLimit = int(endpointRateLimit.Int64)
			}

			authRequired := false
			if endpointAuthRequired.Valid {
				authRequired = endpointAuthRequired.Bool
			}

			timeout := 0
			if endpointTimeout.Valid {
				timeout = int(endpointTimeout.Int64)
			}

			cacheTTL := 0
			if endpointCacheTTL.Valid {
				cacheTTL = int(endpointCacheTTL.Int64)
			}

			endpoint := entity.Endpoint{
				Path:         endpointPath.String,
				Methods:      methods,
				RateLimit:    rateLimit,
				AuthRequired: authRequired,
				Timeout:      timeout,
				CacheTTL:     cacheTTL,
			}
			service.AddEndpoint(endpoint)
		}

		// Add metadata if it exists
		if metadataKey.Valid && metadataValue.Valid {
			service.AddMetadata(metadataKey.String, metadataValue.String)
		}
	}

	// Convert map to slice
	services := make([]*entity.Service, 0, len(servicesMap))
	for _, service := range servicesMap {
		services = append(services, service)
	}

	// Cache the result
	if r.cache != nil {
		err := r.cache.Set(ctx, "services:all", services, 5*time.Minute)
		if err != nil {
			r.logger.Error("Failed to cache services", "error", err)
			// Continue even if caching fails
		}
	}

	r.logger.Info("Retrieved services", "count", len(services))

	return services, nil
}

// GetByID returns a service by its ID
func (r *ServiceRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.Service, error) {
	// Try to get from cache
	if r.cache != nil {
		cachedService, found := r.cache.Get(ctx, fmt.Sprintf("services:%s", id))
		if found {
			r.logger.Debug("Cache hit for service", "id", id)
			return cachedService.(*entity.Service), nil
		}
	}

	// Query database
	query := `
		SELECT 
			s.id, s.name, s.version, s.description, s.base_url, s.timeout, s.retry_count, s.is_active
		FROM services s
		WHERE s.id = ?
	`

	var serviceID, serviceName, serviceVersion, serviceDescription, serviceBaseURL string
	var serviceTimeout, serviceRetryCount int
	var serviceIsActive bool

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&serviceID, &serviceName, &serviceVersion, &serviceDescription, &serviceBaseURL,
		&serviceTimeout, &serviceRetryCount, &serviceIsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("Service not found", "id", id)
			return nil, fmt.Errorf("service not found: %s", id)
		}
		r.logger.Error("Failed to query service", "id", id, "error", err)
		return nil, err
	}

	// Create service
	service := entity.NewService(
		serviceID,
		serviceName,
		serviceVersion,
		serviceDescription,
		serviceBaseURL,
	)
	service.Timeout = serviceTimeout
	service.RetryCount = serviceRetryCount
	service.IsActive = serviceIsActive

	// Query endpoints
	endpointsQuery := `
		SELECT path, methods, rate_limit, auth_required, timeout, cache_ttl
		FROM endpoints
		WHERE service_id = ?
	`

	endpointRows, err := r.db.QueryContext(ctx, endpointsQuery, id)
	if err != nil {
		r.logger.Error("Failed to query endpoints", "service_id", id, "error", err)
		return nil, err
	}
	defer endpointRows.Close()

	// Iterate through endpoint rows
	for endpointRows.Next() {
		var path string
		var methodsStr string
		var rateLimit, timeout, cacheTTL int
		var authRequired bool

		err := endpointRows.Scan(&path, &methodsStr, &rateLimit, &authRequired, &timeout, &cacheTTL)
		if err != nil {
			r.logger.Error("Failed to scan endpoint row", "error", err)
			return nil, err
		}

		methods := strings.Split(methodsStr, ",")

		endpoint := entity.Endpoint{
			Path:         path,
			Methods:      methods,
			RateLimit:    rateLimit,
			AuthRequired: authRequired,
			Timeout:      timeout,
			CacheTTL:     cacheTTL,
		}
		service.AddEndpoint(endpoint)
	}

	// Query metadata
	metadataQuery := `
		SELECT key, value
		FROM service_metadata
		WHERE service_id = ?
	`

	metadataRows, err := r.db.QueryContext(ctx, metadataQuery, id)
	if err != nil {
		r.logger.Error("Failed to query metadata", "service_id", id, "error", err)
		return nil, err
	}
	defer metadataRows.Close()

	// Iterate through metadata rows
	for metadataRows.Next() {
		var key, value string

		err := metadataRows.Scan(&key, &value)
		if err != nil {
			r.logger.Error("Failed to scan metadata row", "error", err)
			return nil, err
		}

		service.AddMetadata(key, value)
	}

	// Cache the result
	if r.cache != nil {
		err := r.cache.Set(ctx, fmt.Sprintf("services:%s", id), service, 5*time.Minute)
		if err != nil {
			r.logger.Error("Failed to cache service", "id", id, "error", err)
			// Continue even if caching fails
		}
	}

	r.logger.Info("Retrieved service", "id", id, "name", service.Name)

	return service, nil
}

// GetByName returns a service by its name
func (r *ServiceRepositoryImpl) GetByName(ctx context.Context, name string) (*entity.Service, error) {
	// Try to get from cache
	if r.cache != nil {
		cachedService, found := r.cache.Get(ctx, fmt.Sprintf("services:name:%s", name))
		if found {
			r.logger.Debug("Cache hit for service by name", "name", name)
			return cachedService.(*entity.Service), nil
		}
	}

	// Query database
	query := `
		SELECT 
			s.id, s.name, s.version, s.description, s.base_url, s.timeout, s.retry_count, s.is_active
		FROM services s
		WHERE s.name = ?
	`

	var serviceID, serviceName, serviceVersion, serviceDescription, serviceBaseURL string
	var serviceTimeout, serviceRetryCount int
	var serviceIsActive bool

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&serviceID, &serviceName, &serviceVersion, &serviceDescription, &serviceBaseURL,
		&serviceTimeout, &serviceRetryCount, &serviceIsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("Service not found by name", "name", name)
			return nil, fmt.Errorf("service not found: %s", name)
		}
		r.logger.Error("Failed to query service by name", "name", name, "error", err)
		return nil, err
	}

	// Get service by ID
	return r.GetByID(ctx, serviceID)
}

// GetByEndpoint returns services that match the given path and method
func (r *ServiceRepositoryImpl) GetByEndpoint(ctx context.Context, path, method string) ([]*entity.Service, error) {
	// Try to get from cache
	cacheKey := fmt.Sprintf("services:endpoint:%s:%s", path, method)
	if r.cache != nil {
		cachedServices, found := r.cache.Get(ctx, cacheKey)
		if found {
			r.logger.Debug("Cache hit for services by endpoint", "path", path, "method", method)
			return cachedServices.([]*entity.Service), nil
		}
	}

	// Query database
	query := `
		SELECT s.id
		FROM services s
		JOIN endpoints e ON s.id = e.service_id
		WHERE e.path = ? AND (e.methods LIKE ? OR e.methods LIKE '%*%')
		AND s.is_active = 1
	`

	rows, err := r.db.QueryContext(ctx, query, path, fmt.Sprintf("%%%s%%", method))
	if err != nil {
		r.logger.Error("Failed to query services by endpoint", "path", path, "method", method, "error", err)
		return nil, err
	}
	defer rows.Close()

	// Collect service IDs
	var serviceIDs []string
	for rows.Next() {
		var serviceID string
		err := rows.Scan(&serviceID)
		if err != nil {
			r.logger.Error("Failed to scan service ID", "error", err)
			return nil, err
		}
		serviceIDs = append(serviceIDs, serviceID)
	}

	// Get services by ID
	services := make([]*entity.Service, 0, len(serviceIDs))
	for _, serviceID := range serviceIDs {
		service, err := r.GetByID(ctx, serviceID)
		if err != nil {
			r.logger.Error("Failed to get service by ID", "id", serviceID, "error", err)
			continue
		}
		services = append(services, service)
	}

	// Cache the result
	if r.cache != nil {
		err := r.cache.Set(ctx, cacheKey, services, 5*time.Minute)
		if err != nil {
			r.logger.Error("Failed to cache services by endpoint", "path", path, "method", method, "error", err)
			// Continue even if caching fails
		}
	}

	r.logger.Info("Retrieved services by endpoint", "path", path, "method", method, "count", len(services))

	return services, nil
}

// Create creates a new service
func (r *ServiceRepositoryImpl) Create(ctx context.Context, service *entity.Service) error {
	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	// Insert service
	query := `
		INSERT INTO services (id, name, version, description, base_url, timeout, retry_count, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = tx.ExecContext(ctx, query,
		service.ID,
		service.Name,
		service.Version,
		service.Description,
		service.BaseURL,
		service.Timeout,
		service.RetryCount,
		service.IsActive,
	)
	if err != nil {
		r.logger.Error("Failed to insert service", "error", err)
		return err
	}

	// Insert endpoints
	for _, endpoint := range service.Endpoints {
		endpointQuery := `
			INSERT INTO endpoints (service_id, path, methods, rate_limit, auth_required, timeout, cache_ttl)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`

		_, err = tx.ExecContext(ctx, endpointQuery,
			service.ID,
			endpoint.Path,
			strings.Join(endpoint.Methods, ","),
			endpoint.RateLimit,
			endpoint.AuthRequired,
			endpoint.Timeout,
			endpoint.CacheTTL,
		)
		if err != nil {
			r.logger.Error("Failed to insert endpoint", "error", err)
			return err
		}
	}

	// Insert metadata
	for key, value := range service.Metadata {
		metadataQuery := `
			INSERT INTO service_metadata (service_id, key, value)
			VALUES (?, ?, ?)
		`

		_, err = tx.ExecContext(ctx, metadataQuery,
			service.ID,
			key,
			value,
		)
		if err != nil {
			r.logger.Error("Failed to insert metadata", "error", err)
			return err
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		r.logger.Error("Failed to commit transaction", "error", err)
		return err
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.Delete(ctx, "services:all")
		r.cache.Delete(ctx, fmt.Sprintf("services:%s", service.ID))
		r.cache.Delete(ctx, fmt.Sprintf("services:name:%s", service.Name))
	}

	r.logger.Info("Created service", "id", service.ID, "name", service.Name)

	return nil
}

// Update updates an existing service
func (r *ServiceRepositoryImpl) Update(ctx context.Context, service *entity.Service) error {
	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	// Update service
	query := `
		UPDATE services
		SET name = ?, version = ?, description = ?, base_url = ?, timeout = ?, retry_count = ?, is_active = ?
		WHERE id = ?
	`

	_, err = tx.ExecContext(ctx, query,
		service.Name,
		service.Version,
		service.Description,
		service.BaseURL,
		service.Timeout,
		service.RetryCount,
		service.IsActive,
		service.ID,
	)
	if err != nil {
		r.logger.Error("Failed to update service", "error", err)
		return err
	}

	// Delete existing endpoints
	_, err = tx.ExecContext(ctx, "DELETE FROM endpoints WHERE service_id = ?", service.ID)
	if err != nil {
		r.logger.Error("Failed to delete endpoints", "error", err)
		return err
	}

	// Insert endpoints
	for _, endpoint := range service.Endpoints {
		endpointQuery := `
			INSERT INTO endpoints (service_id, path, methods, rate_limit, auth_required, timeout, cache_ttl)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`

		_, err = tx.ExecContext(ctx, endpointQuery,
			service.ID,
			endpoint.Path,
			strings.Join(endpoint.Methods, ","),
			endpoint.RateLimit,
			endpoint.AuthRequired,
			endpoint.Timeout,
			endpoint.CacheTTL,
		)
		if err != nil {
			r.logger.Error("Failed to insert endpoint", "error", err)
			return err
		}
	}

	// Delete existing metadata
	_, err = tx.ExecContext(ctx, "DELETE FROM service_metadata WHERE service_id = ?", service.ID)
	if err != nil {
		r.logger.Error("Failed to delete metadata", "error", err)
		return err
	}

	// Insert metadata
	for key, value := range service.Metadata {
		metadataQuery := `
			INSERT INTO service_metadata (service_id, key, value)
			VALUES (?, ?, ?)
		`

		_, err = tx.ExecContext(ctx, metadataQuery,
			service.ID,
			key,
			value,
		)
		if err != nil {
			r.logger.Error("Failed to insert metadata", "error", err)
			return err
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		r.logger.Error("Failed to commit transaction", "error", err)
		return err
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.Delete(ctx, "services:all")
		r.cache.Delete(ctx, fmt.Sprintf("services:%s", service.ID))
		r.cache.Delete(ctx, fmt.Sprintf("services:name:%s", service.Name))
		
		// Invalidate endpoint caches - this is a simplistic approach
		// In a real application, you would track which endpoints are affected
		for _, endpoint := range service.Endpoints {
			for _, method := range endpoint.Methods {
				r.cache.Delete(ctx, fmt.Sprintf("services:endpoint:%s:%s", endpoint.Path, method))
			}
		}
	}

	r.logger.Info("Updated service", "id", service.ID, "name", service.Name)

	return nil
}

// Delete deletes a service by its ID
func (r *ServiceRepositoryImpl) Delete(ctx context.Context, id string) error {
	// Get service first to invalidate caches later
	service, err := r.GetByID(ctx, id)
	if err != nil {
		r.logger.Error("Failed to get service for deletion", "id", id, "error", err)
		return err
	}

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	// Delete metadata
	_, err = tx.ExecContext(ctx, "DELETE FROM service_metadata WHERE service_id = ?", id)
	if err != nil {
		r.logger.Error("Failed to delete metadata", "error", err)
		return err
	}

	// Delete endpoints
	_, err = tx.ExecContext(ctx, "DELETE FROM endpoints WHERE service_id = ?", id)
	if err != nil {
		r.logger.Error("Failed to delete endpoints", "error", err)
		return err
	}

	// Delete service
	_, err = tx.ExecContext(ctx, "DELETE FROM services WHERE id = ?", id)
	if err != nil {
		r.logger.Error("Failed to delete service", "error", err)
		return err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		r.logger.Error("Failed to commit transaction", "error", err)
		return err
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.Delete(ctx, "services:all")
		r.cache.Delete(ctx, fmt.Sprintf("services:%s", id))
		r.cache.Delete(ctx, fmt.Sprintf("services:name:%s", service.Name))
		
		// Invalidate endpoint caches
		for _, endpoint := range service.Endpoints {
			for _, method := range endpoint.Methods {
				r.cache.Delete(ctx, fmt.Sprintf("services:endpoint:%s:%s", endpoint.Path, method))
			}
		}
	}

	r.logger.Info("Deleted service", "id", id, "name", service.Name)

	return nil
}
```

### External Service Clients

External service clients provide implementations for communicating with external services.

#### `http_client.go`

```go
package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"api-gateway/internal/domain/entity"
)

// HTTPClient implements an HTTP client for communicating with backend services
type HTTPClient struct {
	client  *http.Client
	logger  Logger
	metrics MetricsRecorder
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// MetricsRecorder defines the interface for recording metrics
type MetricsRecorder interface {
	RecordServiceLatency(serviceID string, latencyMs int64)
}

// NewHTTPClient creates a new HTTPClient
func NewHTTPClient(timeout time.Duration, logger Logger, metrics MetricsRecorder) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     60 * time.Second,
			},
		},
		logger:  logger,
		metrics: metrics,
	}
}

// SendRequest sends an HTTP request to a backend service
func (c *HTTPClient) SendRequest(ctx context.Context, request *entity.Request, service *entity.Service) (*entity.Response, error) {
	startTime := time.Now()

	// Create target URL
	targetURL := fmt.Sprintf("%s%s", service.BaseURL, request.Path)
	
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, request.Method, targetURL, bytes.NewReader(request.Body))
	if err != nil {
		c.logger.Error("Failed to create HTTP request", "error", err)
		return nil, err
	}

	// Set headers
	for key, values := range request.Headers {
		for _, value := range values {
			httpReq.Header.Add(key, value)
		}
	}

	// Set query parameters
	q := httpReq.URL.Query()
	for key, values := range request.QueryParams {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	httpReq.URL.RawQuery = q.Encode()

	// Set timeout if specified
	if request.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, request.Timeout)
		defer cancel()
		httpReq = httpReq.WithContext(ctx)
	}

	// Send request
	c.logger.Debug("Sending request to backend service",
		"method", request.Method,
		"url", targetURL,
		"service_id", service.ID,
	)

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("Failed to send request to backend service",
			"method", request.Method,
			"url", targetURL,
			"service_id", service.ID,
			"error", err,
		)
		return nil, err
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body",
			"method", request.Method,
			"url", targetURL,
			"service_id", service.ID,
			"error", err,
		)
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

	c.logger.Debug("Received response from backend service",
		"method", request.Method,
		"url", targetURL,
		"service_id", service.ID,
		"status_code", response.StatusCode,
		"latency_ms", response.LatencyMs,
	)

	return response, nil
}

// SetTimeout sets the client timeout
func (c *HTTPClient) SetTimeout(timeout time.Duration) {
	c.client.Timeout = timeout
}
```

### Cache Implementations

Cache implementations provide concrete implementations for caching.

#### `redis_cache.go`

```go
package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"

	"api-gateway/internal/domain/entity"
	"api-gateway/internal/domain/service"
)

// RedisCache implements the CacheService interface using Redis
type RedisCache struct {
	client *redis.Client
	logger Logger
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewRedisCache creates a new RedisCache
func NewRedisCache(client *redis.Client, logger Logger) service.CacheService {
	return &RedisCache{
		client: client,
		logger: logger,
	}
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
		c.logger.Error("Failed to get from Redis", "key", key, "error", err)
		return nil, false, err
	}

	// Unmarshal response
	var response entity.Response
	err = json.Unmarshal(data, &response)
	if err != nil {
		c.logger.Error("Failed to unmarshal response", "key", key, "error", err)
		return nil, false, err
	}

	// Mark as cached
	response.SetCached(true)

	c.logger.Debug("Cache hit", "key", key)

	return &response, true, nil
}

// Set caches a response for a request
func (c *RedisCache) Set(ctx context.Context, request *entity.Request, response *entity.Response, ttl time.Duration) error {
	// Generate cache key
	key := c.generateCacheKey(request)

	// Marshal response
	data, err := json.Marshal(response)
	if err != nil {
		c.logger.Error("Failed to marshal response", "key", key, "error", err)
		return err
	}

	// Set in Redis
	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		c.logger.Error("Failed to set in Redis", "key", key, "error", err)
		return err
	}

	c.logger.Debug("Cache set", "key", key, "ttl", ttl)

	return nil
}

// Delete deletes a cached response for a request
func (c *RedisCache) Delete(ctx context.Context, request *entity.Request) error {
	// Generate cache key
	key := c.generateCacheKey(request)

	// Delete from Redis
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete from Redis", "key", key, "error", err)
		return err
	}

	c.logger.Debug("Cache deleted", "key", key)

	return nil
}

// Clear clears all cached responses
func (c *RedisCache) Clear(ctx context.Context) error {
	// Get all keys with pattern
	keys, err := c.client.Keys(ctx, "cache:*").Result()
	if err != nil {
		c.logger.Error("Failed to get keys from Redis", "error", err)
		return err
	}

	// Delete all keys
	if len(keys) > 0 {
		err = c.client.Del(ctx, keys...).Err()
		if err != nil {
			c.logger.Error("Failed to delete keys from Redis", "error", err)
			return err
		}
	}

	c.logger.Info("Cache cleared", "count", len(keys))

	return nil
}

// generateCacheKey generates a cache key for a request
func (c *RedisCache) generateCacheKey(request *entity.Request) string {
	// Simple implementation - in a real application, you would use a more sophisticated approach
	return "cache:" + request.Method + ":" + request.Path
}

// Get gets a value from the cache
func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, bool) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			// Key not found
			return nil, false
		}
		c.logger.Error("Failed to get from Redis", "key", key, "error", err)
		return nil, false
	}

	var value interface{}
	err = json.Unmarshal(data, &value)
	if err != nil {
		c.logger.Error("Failed to unmarshal value", "key", key, "error", err)
		return nil, false
	}

	return value, true
}

// Set sets a value in the cache
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("Failed to marshal value", "key", key, "error", err)
		return err
	}

	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		c.logger.Error("Failed to set in Redis", "key", key, "error", err)
		return err
	}

	return nil
}

// Delete deletes a value from the cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete from Redis", "key", key, "error", err)
		return err
	}

	return nil
}
```

### Authentication Implementations

Authentication implementations provide concrete implementations for authentication.

#### `jwt_auth.go`

```go
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"api-gateway/internal/domain/entity"
	"api-gateway/internal/domain/service"
)

// JWTAuth implements the AuthService interface using JWT
type JWTAuth struct {
	secretKey     []byte
	issuer        string
	expiration    time.Duration
	logger        Logger
	apiKeyRepo    APIKeyRepository
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// APIKeyRepository defines the interface for API key repository
type APIKeyRepository interface {
	GetByKey(ctx context.Context, key string) (*APIKey, error)
}

// APIKey represents an API key
type APIKey struct {
	Key       string
	ClientID  string
	ExpiresAt int64
	Scopes    []string
}

// NewJWTAuth creates a new JWTAuth
func NewJWTAuth(secretKey []byte, issuer string, expiration time.Duration, logger Logger, apiKeyRepo APIKeyRepository) service.AuthService {
	return &JWTAuth{
		secretKey:  secretKey,
		issuer:     issuer,
		expiration: expiration,
		logger:     logger,
		apiKeyRepo: apiKeyRepo,
	}
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

	a.logger.Debug("No authentication credentials found")
	return false, "", nil
}

// authenticateWithAPIKey authenticates with an API key
func (a *JWTAuth) authenticateWithAPIKey(ctx context.Context, apiKey string) (bool, string, error) {
	if a.apiKeyRepo == nil {
		a.logger.Error("API key repository not configured")
		return false, "", errors.New("API key repository not configured")
	}

	// Get API key from repository
	key, err := a.apiKeyRepo.GetByKey(ctx, apiKey)
	if err != nil {
		a.logger.Error("Failed to get API key", "error", err)
		return false, "", err
	}

	// Check if API key is expired
	if key.ExpiresAt > 0 && time.Now().Unix() > key.ExpiresAt {
		a.logger.Debug("API key expired")
		return false, "", errors.New("API key expired")
	}

	a.logger.Debug("API key authentication successful", "client_id", key.ClientID)
	return true, key.ClientID, nil
}

// authenticateWithJWT authenticates with a JWT token
func (a *JWTAuth) authenticateWithJWT(ctx context.Context, authHeader string) (bool, string, error) {
	// Extract token from header
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		a.logger.Debug("Invalid Authorization header format")
		return false, "", errors.New("invalid Authorization header format")
	}

	// Parse token
	claims, err := a.ValidateToken(ctx, tokenString)
	if err != nil {
		a.logger.Error("Failed to validate token", "error", err)
		return false, "", err
	}

	// Get user ID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		a.logger.Error("Token does not contain user ID")
		return false, "", errors.New("token does not contain user ID")
	}

	a.logger.Debug("JWT authentication successful", "user_id", userID)
	return true, userID, nil
}

// Authorize authorizes a request for a specific service and endpoint
func (a *JWTAuth) Authorize(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) error {
	// If authentication is not required, authorize by default
	if !endpoint.AuthRequired {
		return nil
	}

	// If request is not authenticated, deny access
	if !request.Authenticated {
		a.logger.Debug("Request not authenticated")
		return errors.New("request not authenticated")
	}

	// Check for API key in header
	apiKeyHeader := request.Headers.Get("X-API-Key")
	if apiKeyHeader != "" {
		return a.authorizeWithAPIKey(ctx, apiKeyHeader, service, endpoint)
	}

	// Check for JWT token in Authorization header
	authHeader := request.Headers.Get("Authorization")
	if authHeader != "" {
		return a.authorizeWithJWT(ctx, authHeader, service, endpoint)
	}

	a.logger.Debug("No authorization credentials found")
	return errors.New("no authorization credentials found")
}

// authorizeWithAPIKey authorizes with an API key
func (a *JWTAuth) authorizeWithAPIKey(ctx context.Context, apiKey string, service *entity.Service, endpoint *entity.Endpoint) error {
	if a.apiKeyRepo == nil {
		a.logger.Error("API key repository not configured")
		return errors.New("API key repository not configured")
	}

	// Get API key from repository
	key, err := a.apiKeyRepo.GetByKey(ctx, apiKey)
	if err != nil {
		a.logger.Error("Failed to get API key", "error", err)
		return err
	}

	// Check if API key has required scope
	// This is a simplified implementation - in a real application, you would check against required scopes
	hasScope := false
	for _, scope := range key.Scopes {
		if scope == "*" || scope == service.ID {
			hasScope = true
			break
		}
	}

	if !hasScope {
		a.logger.Debug("API key does not have required scope")
		return errors.New("API key does not have required scope")
	}

	a.logger.Debug("API key authorization successful")
	return nil
}

// authorizeWithJWT authorizes with a JWT token
func (a *JWTAuth) authorizeWithJWT(ctx context.Context, authHeader string, service *entity.Service, endpoint *entity.Endpoint) error {
	// Extract token from header
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		a.logger.Debug("Invalid Authorization header format")
		return errors.New("invalid Authorization header format")
	}

	// Parse token
	claims, err := a.ValidateToken(ctx, tokenString)
	if err != nil {
		a.logger.Error("Failed to validate token", "error", err)
		return err
	}

	// Check if token has required scope
	// This is a simplified implementation - in a real application, you would check against required scopes
	scope, ok := claims["scope"].(string)
	if !ok {
		a.logger.Debug("Token does not contain scope")
		return errors.New("token does not contain scope")
	}

	if scope != "*" && scope != service.ID {
		a.logger.Debug("Token does not have required scope")
		return errors.New("token does not have required scope")
	}

	a.logger.Debug("JWT authorization successful")
	return nil
}

// GenerateToken generates an authentication token
func (a *JWTAuth) GenerateToken(ctx context.Context, userID string, claims map[string]interface{}) (string, error) {
	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	tokenClaims := token.Claims.(jwt.MapClaims)
	tokenClaims["sub"] = userID
	tokenClaims["iss"] = a.issuer
	tokenClaims["iat"] = time.Now().Unix()
	tokenClaims["exp"] = time.Now().Add(a.expiration).Unix()

	// Add custom claims
	for key, value := range claims {
		tokenClaims[key] = value
	}

	// Sign token
	tokenString, err := token.SignedString(a.secretKey)
	if err != nil {
		a.logger.Error("Failed to sign token", "error", err)
		return "", err
	}

	a.logger.Debug("Generated token", "user_id", userID)
	return tokenString, nil
}

// ValidateToken validates an authentication token
func (a *JWTAuth) ValidateToken(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.secretKey, nil
	})
	if err != nil {
		a.logger.Error("Failed to parse token", "error", err)
		return nil, err
	}

	// Validate token
	if !token.Valid {
		a.logger.Debug("Invalid token")
		return nil, errors.New("invalid token")
	}

	// Get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		a.logger.Error("Failed to get claims from token")
		return nil, errors.New("failed to get claims from token")
	}

	// Validate issuer
	if claims["iss"] != a.issuer {
		a.logger.Debug("Invalid token issuer")
		return nil, errors.New("invalid token issuer")
	}

	// Validate expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		a.logger.Debug("Token does not contain expiration")
		return nil, errors.New("token does not contain expiration")
	}
	if time.Now().Unix() > int64(exp) {
		a.logger.Debug("Token expired")
		return nil, errors.New("token expired")
	}

	// Convert claims to map
	result := make(map[string]interface{})
	for key, value := range claims {
		result[key] = value
	}

	a.logger.Debug("Token validated")
	return result, nil
}
```

### Rate Limiting Implementations

Rate limiting implementations provide concrete implementations for rate limiting.

#### `token_bucket_rate_limiter.go`

```go
package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"api-gateway/internal/domain/entity"
	"api-gateway/internal/domain/service"
)

// TokenBucketRateLimiter implements the RateLimitService interface using the token bucket algorithm
type TokenBucketRateLimiter struct {
	redis  *redis.Client
	logger Logger
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewTokenBucketRateLimiter creates a new TokenBucketRateLimiter
func NewTokenBucketRateLimiter(redis *redis.Client, logger Logger) service.RateLimitService {
	return &TokenBucketRateLimiter{
		redis:  redis,
		logger: logger,
	}
}

// CheckLimit checks if a request exceeds the rate limit
func (r *TokenBucketRateLimiter) CheckLimit(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) (bool, error) {
	// If rate limit is not set, allow the request
	if endpoint.RateLimit <= 0 {
		return true, nil
	}

	// Generate key
	key := r.generateKey(request, service, endpoint)

	// Get current tokens
	tokens, err := r.getCurrentTokens(ctx, key, endpoint.RateLimit)
	if err != nil {
		r.logger.Error("Failed to get current tokens", "key", key, "error", err)
		// Allow the request if there's an error
		return true, err
	}

	// Check if there are tokens available
	if tokens <= 0 {
		r.logger.Debug("Rate limit exceeded", "key", key)
		return false, nil
	}

	r.logger.Debug("Rate limit check passed", "key", key, "tokens", tokens)
	return true, nil
}

// RecordRequest records a request for rate limiting purposes
func (r *TokenBucketRateLimiter) RecordRequest(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) error {
	// If rate limit is not set, do nothing
	if endpoint.RateLimit <= 0 {
		return nil
	}

	// Generate key
	key := r.generateKey(request, service, endpoint)

	// Consume a token
	err := r.consumeToken(ctx, key)
	if err != nil {
		r.logger.Error("Failed to consume token", "key", key, "error", err)
		return err
	}

	r.logger.Debug("Request recorded for rate limiting", "key", key)
	return nil
}

// GetLimit gets the current rate limit for a client
func (r *TokenBucketRateLimiter) GetLimit(ctx context.Context, clientID string, service *entity.Service, endpoint *entity.Endpoint) (int, int, error) {
	// If rate limit is not set, return unlimited
	if endpoint.RateLimit <= 0 {
		return 0, 0, nil
	}

	// Generate key
	key := fmt.Sprintf("rate_limit:%s:%s:%s", clientID, service.ID, endpoint.Path)

	// Get current tokens
	tokens, err := r.getCurrentTokens(ctx, key, endpoint.RateLimit)
	if err != nil {
		r.logger.Error("Failed to get current tokens", "key", key, "error", err)
		return 0, 0, err
	}

	r.logger.Debug("Got rate limit", "key", key, "limit", endpoint.RateLimit, "remaining", tokens)
	return endpoint.RateLimit, tokens, nil
}

// generateKey generates a key for rate limiting
func (r *TokenBucketRateLimiter) generateKey(request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) string {
	// Use client ID or IP address as identifier
	identifier := request.UserID
	if identifier == "" {
		identifier = request.ClientIP
	}

	return fmt.Sprintf("rate_limit:%s:%s:%s", identifier, service.ID, endpoint.Path)
}

// getCurrentTokens gets the current number of tokens for a key
func (r *TokenBucketRateLimiter) getCurrentTokens(ctx context.Context, key string, limit int) (int, error) {
	// Check if key exists
	exists, err := r.redis.Exists(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// If key doesn't exist, initialize it
	if exists == 0 {
		// Set initial tokens to limit
		err = r.redis.Set(ctx, key, limit, time.Minute).Err()
		if err != nil {
			return 0, err
		}
		return limit, nil
	}

	// Get current tokens
	tokens, err := r.redis.Get(ctx, key).Int()
	if err != nil {
		return 0, err
	}

	return tokens, nil
}

// consumeToken consumes a token for a key
func (r *TokenBucketRateLimiter) consumeToken(ctx context.Context, key string) error {
	// Decrement tokens
	count, err := r.redis.Decr(ctx, key).Result()
	if err != nil {
		return err
	}

	// Ensure tokens don't go below 0
	if count < 0 {
		r.redis.Set(ctx, key, 0, time.Minute)
	}

	return nil
}
```

### Database Connections

Database connections provide implementations for connecting to databases.

#### `database.go`

```go
package persistence

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DatabaseConfig represents the configuration for a database connection
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// NewDatabase creates a new database connection
func NewDatabase(config DatabaseConfig) (*sql.DB, error) {
	var dsn string

	switch config.Driver {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			config.User, config.Password, config.Host, config.Port, config.Database)
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)
	case "sqlite3":
		dsn = config.Database
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	db, err := sql.Open(config.Driver, dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// InitSchema initializes the database schema
func InitSchema(db *sql.DB) error {
	// Create services table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS services (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			version VARCHAR(50) NOT NULL,
			description TEXT,
			base_url VARCHAR(255) NOT NULL,
			timeout INT NOT NULL,
			retry_count INT NOT NULL,
			is_active BOOLEAN NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create endpoints table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS endpoints (
			id VARCHAR(36) PRIMARY KEY,
			service_id VARCHAR(36) NOT NULL,
			path VARCHAR(255) NOT NULL,
			methods VARCHAR(255) NOT NULL,
			rate_limit INT NOT NULL,
			auth_required BOOLEAN NOT NULL,
			timeout INT NOT NULL,
			cache_ttl INT NOT NULL,
			FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create service_metadata table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS service_metadata (
			id VARCHAR(36) PRIMARY KEY,
			service_id VARCHAR(36) NOT NULL,
			key VARCHAR(255) NOT NULL,
			value TEXT NOT NULL,
			FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create api_keys table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS api_keys (
			key VARCHAR(255) PRIMARY KEY,
			client_id VARCHAR(255) NOT NULL,
			expires_at BIGINT NOT NULL,
			scopes TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	return nil
}
```

### Logging Implementations

Logging implementations provide concrete implementations for logging.

#### `zap_logger.go`

```go
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger implements the Logger interface using zap
type ZapLogger struct {
	logger *zap.SugaredLogger
}

// NewZapLogger creates a new ZapLogger
func NewZapLogger(level string, development bool) (*ZapLogger, error) {
	// Parse log level
	var zapLevel zapcore.Level
	err := zapLevel.UnmarshalText([]byte(level))
	if err != nil {
		zapLevel = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create core
	var core zapcore.Core
	if development {
		// Development mode - pretty console output
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapLevel,
		)
	} else {
		// Production mode - JSON output
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapLevel,
		)
	}

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	sugaredLogger := logger.Sugar()

	return &ZapLogger{
		logger: sugaredLogger,
	}, nil
}

// Info logs an info message
func (l *ZapLogger) Info(msg string, fields ...interface{}) {
	l.logger.Infow(msg, fields...)
}

// Error logs an error message
func (l *ZapLogger) Error(msg string, fields ...interface{}) {
	l.logger.Errorw(msg, fields...)
}

// Debug logs a debug message
func (l *ZapLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debugw(msg, fields...)
}

// Warn logs a warning message
func (l *ZapLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Warnw(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *ZapLogger) Fatal(msg string, fields ...interface{}) {
	l.logger.Fatalw(msg, fields...)
}

// With returns a logger with additional fields
func (l *ZapLogger) With(fields ...interface{}) *ZapLogger {
	return &ZapLogger{
		logger: l.logger.With(fields...),
	}
}
```

### Configuration Implementations

Configuration implementations provide concrete implementations for loading configuration.

#### `config.go`

```go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Logging  LoggingConfig
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// RedisConfig represents the Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// AuthConfig represents the authentication configuration
type AuthConfig struct {
	SecretKey  string
	Issuer     string
	Expiration time.Duration
}

// LoggingConfig represents the logging configuration
type LoggingConfig struct {
	Level       string
	Development bool
}

// LoadConfig loads the application configuration
func LoadConfig(configPath string) (*Config, error) {
	// Set defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "5s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.shutdown_timeout", "30s")

	viper.SetDefault("database.driver", "sqlite3")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.user", "root")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.database", "api_gateway")
	viper.SetDefault("database.ssl_mode", "disable")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("auth.secret_key", "secret")
	viper.SetDefault("auth.issuer", "api-gateway")
	viper.SetDefault("auth.expiration", "1h")

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.development", true)

	// Set config file
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.SetConfigName("config")
	}

	// Set environment variable prefix
	viper.SetEnvPrefix("API_GATEWAY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// Config file not found, use defaults and environment variables
	}

	// Parse config
	var config Config

	// Server config
	config.Server.Port = viper.GetInt("server.port")
	readTimeout, err := time.ParseDuration(viper.GetString("server.read_timeout"))
	if err != nil {
		return nil, fmt.Errorf("invalid read timeout: %w", err)
	}
	config.Server.ReadTimeout = readTimeout

	writeTimeout, err := time.ParseDuration(viper.GetString("server.write_timeout"))
	if err != nil {
		return nil, fmt.Errorf("invalid write timeout: %w", err)
	}
	config.Server.WriteTimeout = writeTimeout

	shutdownTimeout, err := time.ParseDuration(viper.GetString("server.shutdown_timeout"))
	if err != nil {
		return nil, fmt.Errorf("invalid shutdown timeout: %w", err)
	}
	config.Server.ShutdownTimeout = shutdownTimeout

	// Database config
	config.Database.Driver = viper.GetString("database.driver")
	config.Database.Host = viper.GetString("database.host")
	config.Database.Port = viper.GetInt("database.port")
	config.Database.User = viper.GetString("database.user")
	config.Database.Password = viper.GetString("database.password")
	config.Database.Database = viper.GetString("database.database")
	config.Database.SSLMode = viper.GetString("database.ssl_mode")

	// Redis config
	config.Redis.Host = viper.GetString("redis.host")
	config.Redis.Port = viper.GetInt("redis.port")
	config.Redis.Password = viper.GetString("redis.password")
	config.Redis.DB = viper.GetInt("redis.db")

	// Auth config
	config.Auth.SecretKey = viper.GetString("auth.secret_key")
	config.Auth.Issuer = viper.GetString("auth.issuer")
	expiration, err := time.ParseDuration(viper.GetString("auth.expiration"))
	if err != nil {
		return nil, fmt.Errorf("invalid auth expiration: %w", err)
	}
	config.Auth.Expiration = expiration

	// Logging config
	config.Logging.Level = viper.GetString("logging.level")
	config.Logging.Development = viper.GetBool("logging.development")

	// Override with environment variables
	overrideWithEnv(&config)

	return &config, nil
}

// overrideWithEnv overrides config values with environment variables
func overrideWithEnv(config *Config) {
	// Server config
	if port := os.Getenv("API_GATEWAY_SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	if readTimeout := os.Getenv("API_GATEWAY_SERVER_READ_TIMEOUT"); readTimeout != "" {
		if rt, err := time.ParseDuration(readTimeout); err == nil {
			config.Server.ReadTimeout = rt
		}
	}

	if writeTimeout := os.Getenv("API_GATEWAY_SERVER_WRITE_TIMEOUT"); writeTimeout != "" {
		if wt, err := time.ParseDuration(writeTimeout); err == nil {
			config.Server.WriteTimeout = wt
		}
	}

	if shutdownTimeout := os.Getenv("API_GATEWAY_SERVER_SHUTDOWN_TIMEOUT"); shutdownTimeout != "" {
		if st, err := time.ParseDuration(shutdownTimeout); err == nil {
			config.Server.ShutdownTimeout = st
		}
	}

	// Database config
	if driver := os.Getenv("API_GATEWAY_DATABASE_DRIVER"); driver != "" {
		config.Database.Driver = driver
	}

	if host := os.Getenv("API_GATEWAY_DATABASE_HOST"); host != "" {
		config.Database.Host = host
	}

	if port := os.Getenv("API_GATEWAY_DATABASE_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	}

	if user := os.Getenv("API_GATEWAY_DATABASE_USER"); user != "" {
		config.Database.User = user
	}

	if password := os.Getenv("API_GATEWAY_DATABASE_PASSWORD"); password != "" {
		config.Database.Password = password
	}

	if database := os.Getenv("API_GATEWAY_DATABASE_DATABASE"); database != "" {
		config.Database.Database = database
	}

	if sslMode := os.Getenv("API_GATEWAY_DATABASE_SSL_MODE"); sslMode != "" {
		config.Database.SSLMode = sslMode
	}

	// Redis config
	if host := os.Getenv("API_GATEWAY_REDIS_HOST"); host != "" {
		config.Redis.Host = host
	}

	if port := os.Getenv("API_GATEWAY_REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Redis.Port = p
		}
	}

	if password := os.Getenv("API_GATEWAY_REDIS_PASSWORD"); password != "" {
		config.Redis.Password = password
	}

	if db := os.Getenv("API_GATEWAY_REDIS_DB"); db != "" {
		if d, err := strconv.Atoi(db); err == nil {
			config.Redis.DB = d
		}
	}

	// Auth config
	if secretKey := os.Getenv("API_GATEWAY_AUTH_SECRET_KEY"); secretKey != "" {
		config.Auth.SecretKey = secretKey
	}

	if issuer := os.Getenv("API_GATEWAY_AUTH_ISSUER"); issuer != "" {
		config.Auth.Issuer = issuer
	}

	if expiration := os.Getenv("API_GATEWAY_AUTH_EXPIRATION"); expiration != "" {
		if exp, err := time.ParseDuration(expiration); err == nil {
			config.Auth.Expiration = exp
		}
	}

	// Logging config
	if level := os.Getenv("API_GATEWAY_LOGGING_LEVEL"); level != "" {
		config.Logging.Level = level
	}

	if development := os.Getenv("API_GATEWAY_LOGGING_DEVELOPMENT"); development != "" {
		if dev, err := strconv.ParseBool(development); err == nil {
			config.Logging.Development = dev
		}
	}
}
```

### Metrics Implementations

Metrics implementations provide concrete implementations for collecting and reporting metrics.

#### `prometheus_metrics.go`

```go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics implements the MetricsAggregator interface using Prometheus
type PrometheusMetrics struct {
	requestCounter      *prometheus.CounterVec
	requestLatency      *prometheus.HistogramVec
	errorCounter        *prometheus.CounterVec
	serviceLatency      *prometheus.HistogramVec
	statusCodeCounter   *prometheus.CounterVec
}

// NewPrometheusMetrics creates a new PrometheusMetrics
func NewPrometheusMetrics() *PrometheusMetrics {
	requestCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_gateway_requests_total",
			Help: "Total number of requests",
		},
		[]string{"path", "method"},
	)

	requestLatency := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_gateway_request_latency_ms",
			Help:    "Request latency in milliseconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1ms to 512ms
		},
		[]string{"path", "method"},
	)

	errorCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_gateway_errors_total",
			Help: "Total number of errors",
		},
		[]string{"path", "method", "error_type"},
	)

	serviceLatency := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_gateway_service_latency_ms",
			Help:    "Service latency in milliseconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1ms to 512ms
		},
		[]string{"service_id"},
	)

	statusCodeCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_gateway_status_codes_total",
			Help: "Total number of status codes",
		},
		[]string{"path", "method", "status_code"},
	)

	return &PrometheusMetrics{
		requestCounter:    requestCounter,
		requestLatency:    requestLatency,
		errorCounter:      errorCounter,
		serviceLatency:    serviceLatency,
		statusCodeCounter: statusCodeCounter,
	}
}

// RecordRequest records a request for metrics
func (m *PrometheusMetrics) RecordRequest(path string, method string, statusCode int, latencyMs int64) {
	m.requestCounter.WithLabelValues(path, method).Inc()
	m.requestLatency.WithLabelValues(path, method).Observe(float64(latencyMs))
	m.statusCodeCounter.WithLabelValues(path, method, fmt.Sprintf("%d", statusCode)).Inc()
}

// RecordError records an error for metrics
func (m *PrometheusMetrics) RecordError(path string, method string, errorType string) {
	m.errorCounter.WithLabelValues(path, method, errorType).Inc()
}

// RecordServiceLatency records service latency for metrics
func (m *PrometheusMetrics) RecordServiceLatency(serviceID string, latencyMs int64) {
	m.serviceLatency.WithLabelValues(serviceID).Observe(float64(latencyMs))
}
```

## Conclusion

The infrastructure layer provides concrete implementations of the interfaces defined in the domain layer. It includes implementations for repositories, external service clients, authentication, caching, and other infrastructure concerns.

Key components of the infrastructure layer include:

1. **Repository Implementations**: Provide concrete implementations of the repository interfaces defined in the domain layer.
2. **External Service Clients**: Provide implementations for communicating with external services.
3. **Cache Implementations**: Provide concrete implementations for caching.
4. **Authentication Implementations**: Provide concrete implementations for authentication.
5. **Rate Limiting Implementations**: Provide concrete implementations for rate limiting.
6. **Database Connections**: Provide implementations for connecting to databases.
7. **Logging Implementations**: Provide concrete implementations for logging.
8. **Configuration Implementations**: Provide concrete implementations for loading configuration.
9. **Metrics Implementations**: Provide concrete implementations for collecting and reporting metrics.

The infrastructure layer is the outermost layer of the Clean Architecture and depends on the domain and application layers. It provides the concrete implementations that allow the application to interact with external systems and services.
