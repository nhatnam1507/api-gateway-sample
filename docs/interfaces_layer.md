# Interfaces Layer Implementation

This document outlines the implementation details for the interfaces layer of the API Gateway following Clean Architecture principles.

## Overview

The interfaces layer is the entry point to the application and handles the communication between the outside world and the application. It includes HTTP handlers, middleware, and router implementations that connect the application to the outside world.

## Core Components

### API Handlers

API handlers handle HTTP requests and responses.

#### `handler.go`

```go
package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"api-gateway/internal/application/dto"
	"api-gateway/internal/application/usecase"
	"api-gateway/internal/interfaces/dto"
)

// Handler handles HTTP requests
type Handler struct {
	proxyUseCase            *usecase.ProxyUseCase
	authUseCase             *usecase.AuthUseCase
	rateLimitUseCase        *usecase.RateLimitUseCase
	serviceManagementUseCase *usecase.ServiceManagementUseCase
	logger                  Logger
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewHandler creates a new Handler
func NewHandler(
	proxyUseCase *usecase.ProxyUseCase,
	authUseCase *usecase.AuthUseCase,
	rateLimitUseCase *usecase.RateLimitUseCase,
	serviceManagementUseCase *usecase.ServiceManagementUseCase,
	logger Logger,
) *Handler {
	return &Handler{
		proxyUseCase:            proxyUseCase,
		authUseCase:             authUseCase,
		rateLimitUseCase:        rateLimitUseCase,
		serviceManagementUseCase: serviceManagementUseCase,
		logger:                  logger,
	}
}

// ProxyHandler handles proxy requests
func (h *Handler) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", "error", err)
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
		h.logger.Error("Failed to proxy request", "error", err)
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
	_, err = w.Write(responseDTO.Body)
	if err != nil {
		h.logger.Error("Failed to write response body", "error", err)
		return
	}

	// Log request
	h.logger.Info("Proxied request",
		"method", r.Method,
		"path", r.URL.Path,
		"status_code", responseDTO.StatusCode,
		"latency_ms", time.Since(startTime).Milliseconds(),
	)
}

// AuthHandler handles authentication requests
func (h *Handler) AuthHandler(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse request
	var authRequest usecase.AuthRequest
	err = json.Unmarshal(body, &authRequest)
	if err != nil {
		h.logger.Error("Failed to parse auth request", "error", err)
		http.Error(w, "Failed to parse auth request", http.StatusBadRequest)
		return
	}

	// Authenticate
	authResponse, err := h.authUseCase.Authenticate(r.Context(), &authRequest)
	if err != nil {
		h.logger.Error("Authentication failed", "error", err)
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Marshal response
	responseBody, err := json.Marshal(authResponse)
	if err != nil {
		h.logger.Error("Failed to marshal auth response", "error", err)
		http.Error(w, "Failed to marshal auth response", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Write response
	_, err = w.Write(responseBody)
	if err != nil {
		h.logger.Error("Failed to write response body", "error", err)
		return
	}

	h.logger.Info("Authentication successful", "client_id", authRequest.ClientID)
}

// ValidateTokenHandler handles token validation requests
func (h *Handler) ValidateTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		h.logger.Error("Invalid Authorization header")
		http.Error(w, "Invalid Authorization header", http.StatusBadRequest)
		return
	}

	token := authHeader[7:]

	// Validate token
	claims, err := h.authUseCase.ValidateToken(r.Context(), token)
	if err != nil {
		h.logger.Error("Token validation failed", "error", err)
		http.Error(w, "Token validation failed", http.StatusUnauthorized)
		return
	}

	// Marshal response
	responseBody, err := json.Marshal(claims)
	if err != nil {
		h.logger.Error("Failed to marshal claims", "error", err)
		http.Error(w, "Failed to marshal claims", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Write response
	_, err = w.Write(responseBody)
	if err != nil {
		h.logger.Error("Failed to write response body", "error", err)
		return
	}

	h.logger.Info("Token validation successful", "user_id", claims["sub"])
}

// GetServicesHandler handles get services requests
func (h *Handler) GetServicesHandler(w http.ResponseWriter, r *http.Request) {
	// Get all services
	services, err := h.serviceManagementUseCase.GetAllServices(r.Context())
	if err != nil {
		h.logger.Error("Failed to get services", "error", err)
		http.Error(w, "Failed to get services", http.StatusInternalServerError)
		return
	}

	// Marshal response
	responseBody, err := json.Marshal(services)
	if err != nil {
		h.logger.Error("Failed to marshal services", "error", err)
		http.Error(w, "Failed to marshal services", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Write response
	_, err = w.Write(responseBody)
	if err != nil {
		h.logger.Error("Failed to write response body", "error", err)
		return
	}

	h.logger.Info("Get services successful", "count", len(services))
}

// GetServiceHandler handles get service requests
func (h *Handler) GetServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Get service ID from URL
	serviceID := r.URL.Query().Get("id")
	if serviceID == "" {
		h.logger.Error("Missing service ID")
		http.Error(w, "Missing service ID", http.StatusBadRequest)
		return
	}

	// Get service
	service, err := h.serviceManagementUseCase.GetServiceByID(r.Context(), serviceID)
	if err != nil {
		h.logger.Error("Failed to get service", "id", serviceID, "error", err)
		http.Error(w, "Failed to get service", http.StatusNotFound)
		return
	}

	// Marshal response
	responseBody, err := json.Marshal(service)
	if err != nil {
		h.logger.Error("Failed to marshal service", "error", err)
		http.Error(w, "Failed to marshal service", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Write response
	_, err = w.Write(responseBody)
	if err != nil {
		h.logger.Error("Failed to write response body", "error", err)
		return
	}

	h.logger.Info("Get service successful", "id", serviceID)
}

// CreateServiceHandler handles create service requests
func (h *Handler) CreateServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse request
	var serviceDTO dto.ServiceDTO
	err = json.Unmarshal(body, &serviceDTO)
	if err != nil {
		h.logger.Error("Failed to parse service", "error", err)
		http.Error(w, "Failed to parse service", http.StatusBadRequest)
		return
	}

	// Create service
	createdService, err := h.serviceManagementUseCase.CreateService(r.Context(), &serviceDTO)
	if err != nil {
		h.logger.Error("Failed to create service", "error", err)
		http.Error(w, "Failed to create service", http.StatusInternalServerError)
		return
	}

	// Marshal response
	responseBody, err := json.Marshal(createdService)
	if err != nil {
		h.logger.Error("Failed to marshal service", "error", err)
		http.Error(w, "Failed to marshal service", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Write response
	_, err = w.Write(responseBody)
	if err != nil {
		h.logger.Error("Failed to write response body", "error", err)
		return
	}

	h.logger.Info("Create service successful", "id", createdService.ID)
}

// UpdateServiceHandler handles update service requests
func (h *Handler) UpdateServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse request
	var serviceDTO dto.ServiceDTO
	err = json.Unmarshal(body, &serviceDTO)
	if err != nil {
		h.logger.Error("Failed to parse service", "error", err)
		http.Error(w, "Failed to parse service", http.StatusBadRequest)
		return
	}

	// Update service
	updatedService, err := h.serviceManagementUseCase.UpdateService(r.Context(), &serviceDTO)
	if err != nil {
		h.logger.Error("Failed to update service", "error", err)
		http.Error(w, "Failed to update service", http.StatusInternalServerError)
		return
	}

	// Marshal response
	responseBody, err := json.Marshal(updatedService)
	if err != nil {
		h.logger.Error("Failed to marshal service", "error", err)
		http.Error(w, "Failed to marshal service", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Write response
	_, err = w.Write(responseBody)
	if err != nil {
		h.logger.Error("Failed to write response body", "error", err)
		return
	}

	h.logger.Info("Update service successful", "id", updatedService.ID)
}

// DeleteServiceHandler handles delete service requests
func (h *Handler) DeleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Get service ID from URL
	serviceID := r.URL.Query().Get("id")
	if serviceID == "" {
		h.logger.Error("Missing service ID")
		http.Error(w, "Missing service ID", http.StatusBadRequest)
		return
	}

	// Delete service
	err := h.serviceManagementUseCase.DeleteService(r.Context(), serviceID)
	if err != nil {
		h.logger.Error("Failed to delete service", "id", serviceID, "error", err)
		http.Error(w, "Failed to delete service", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.WriteHeader(http.StatusNoContent)

	h.logger.Info("Delete service successful", "id", serviceID)
}

// CheckRateLimitHandler handles check rate limit requests
func (h *Handler) CheckRateLimitHandler(w http.ResponseWriter, r *http.Request) {
	// Get parameters from URL
	clientID := r.URL.Query().Get("client_id")
	path := r.URL.Query().Get("path")
	method := r.URL.Query().Get("method")

	if clientID == "" || path == "" || method == "" {
		h.logger.Error("Missing parameters")
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	// Check rate limit
	rateLimitInfo, err := h.rateLimitUseCase.CheckRateLimit(r.Context(), clientID, path, method)
	if err != nil {
		h.logger.Error("Failed to check rate limit", "error", err)
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Marshal response
	responseBody, err := json.Marshal(rateLimitInfo)
	if err != nil {
		h.logger.Error("Failed to marshal rate limit info", "error", err)
		http.Error(w, "Failed to marshal rate limit info", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimitInfo.Limit))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rateLimitInfo.Remaining))
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", rateLimitInfo.Reset))

	// Write response
	_, err = w.Write(responseBody)
	if err != nil {
		h.logger.Error("Failed to write response body", "error", err)
		return
	}

	h.logger.Info("Check rate limit successful", "client_id", clientID, "path", path, "method", method)
}

// HealthCheckHandler handles health check requests
func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create response
	response := map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}

	// Marshal response
	responseBody, err := json.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal health check response", "error", err)
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Write response
	_, err = w.Write(responseBody)
	if err != nil {
		h.logger.Error("Failed to write response body", "error", err)
		return
	}

	h.logger.Debug("Health check successful")
}

// getClientIP gets the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, use the first one
		ips := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Use RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
```

### Middleware

Middleware handles cross-cutting concerns like logging, authentication, and rate limiting.

#### `middleware.go`

```go
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Middleware represents a middleware function
type Middleware func(http.Handler) http.Handler

// Chain chains multiple middleware together
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

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

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(logger Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered",
						"error", err,
						"path", r.URL.Path,
						"method", r.Method,
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
			
			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// TimeoutMiddleware adds a timeout to the request context
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// MetricsMiddleware collects metrics
func MetricsMiddleware(metrics MetricsRecorder) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			
			// Create a response writer wrapper to capture status code
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			
			// Call next handler
			next.ServeHTTP(rw, r)
			
			// Record metrics
			metrics.RecordRequest(r.URL.Path, r.Method, rw.statusCode, time.Since(startTime).Milliseconds())
		})
	}
}

// AuthMiddleware handles authentication
func AuthMiddleware(authUseCase *usecase.AuthUseCase, logger Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for certain paths
			if r.URL.Path == "/auth" || r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}
			
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
				// Check for API key
				apiKey := r.Header.Get("X-API-Key")
				if apiKey == "" {
					logger.Error("Missing authentication")
					http.Error(w, "Missing authentication", http.StatusUnauthorized)
					return
				}
				
				// API key authentication will be handled by the proxy use case
				next.ServeHTTP(w, r)
				return
			}
			
			token := authHeader[7:]
			
			// Validate token
			claims, err := authUseCase.ValidateToken(r.Context(), token)
			if err != nil {
				logger.Error("Invalid token", "error", err)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			
			// Add claims to request context
			ctx := context.WithValue(r.Context(), "claims", claims)
			r = r.WithContext(ctx)
			
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware handles rate limiting
func RateLimitMiddleware(rateLimitUseCase *usecase.RateLimitUseCase, logger Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting for certain paths
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}
			
			// Get client ID from context or use IP address
			var clientID string
			claims, ok := r.Context().Value("claims").(map[string]interface{})
			if ok {
				if sub, ok := claims["sub"].(string); ok {
					clientID = sub
				}
			}
			
			if clientID == "" {
				clientID = getClientIP(r)
			}
			
			// Check rate limit
			rateLimitInfo, err := rateLimitUseCase.CheckRateLimit(r.Context(), clientID, r.URL.Path, r.Method)
			if err != nil {
				logger.Error("Rate limit exceeded", "client_id", clientID, "path", r.URL.Path, "method", r.Method)
				
				// Set rate limit headers
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimitInfo.Limit))
				w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rateLimitInfo.Remaining))
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", rateLimitInfo.Reset))
				
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimitInfo.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rateLimitInfo.Remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", rateLimitInfo.Reset))
			
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// MetricsRecorder defines the interface for recording metrics
type MetricsRecorder interface {
	RecordRequest(path string, method string, statusCode int, latencyMs int64)
}
```

### Router

The router handles request routing.

#### `router.go`

```go
package api

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Router handles request routing
type Router struct {
	handler    *Handler
	logger     Logger
	metrics    MetricsRecorder
	authUseCase *usecase.AuthUseCase
	rateLimitUseCase *usecase.RateLimitUseCase
}

// NewRouter creates a new Router
func NewRouter(
	handler *Handler,
	logger Logger,
	metrics MetricsRecorder,
	authUseCase *usecase.AuthUseCase,
	rateLimitUseCase *usecase.RateLimitUseCase,
) *Router {
	return &Router{
		handler:    handler,
		logger:     logger,
		metrics:    metrics,
		authUseCase: authUseCase,
		rateLimitUseCase: rateLimitUseCase,
	}
}

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

	// Not found handler
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.logger.Error("Route not found", "path", r.URL.Path, "method", r.Method)
		http.Error(w, "Route not found", http.StatusNotFound)
	})

	// Method not allowed handler
	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.logger.Error("Method not allowed", "path", r.URL.Path, "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	return router
}
```

### Interface DTOs

Interface DTOs define the data transfer objects for the interfaces layer.

#### `api_dto.go`

```go
package dto

import (
	"time"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// HealthCheckResponse represents a health check response
type HealthCheckResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

// RateLimitResponse represents a rate limit response
type RateLimitResponse struct {
	Limit     int   `json:"limit"`
	Remaining int   `json:"remaining"`
	Reset     int64 `json:"reset"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// TokenRequest represents a token request
type TokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	Scope        string `json:"scope"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// ServiceResponse represents a service response
type ServiceResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	BaseURL     string                 `json:"base_url"`
	Timeout     int                    `json:"timeout"`
	RetryCount  int                    `json:"retry_count"`
	IsActive    bool                   `json:"is_active"`
	Endpoints   []EndpointResponse     `json:"endpoints"`
	Metadata    map[string]string      `json:"metadata"`
}

// EndpointResponse represents an endpoint response
type EndpointResponse struct {
	Path         string   `json:"path"`
	Methods      []string `json:"methods"`
	RateLimit    int      `json:"rate_limit"`
	AuthRequired bool     `json:"auth_required"`
	Timeout      int      `json:"timeout"`
	CacheTTL     int      `json:"cache_ttl"`
}

// ServiceRequest represents a service request
type ServiceRequest struct {
	Name        string                `json:"name"`
	Version     string                `json:"version"`
	Description string                `json:"description"`
	BaseURL     string                `json:"base_url"`
	Timeout     int                   `json:"timeout"`
	RetryCount  int                   `json:"retry_count"`
	IsActive    bool                  `json:"is_active"`
	Endpoints   []EndpointRequest     `json:"endpoints"`
	Metadata    map[string]string     `json:"metadata"`
}

// EndpointRequest represents an endpoint request
type EndpointRequest struct {
	Path         string   `json:"path"`
	Methods      []string `json:"methods"`
	RateLimit    int      `json:"rate_limit"`
	AuthRequired bool     `json:"auth_required"`
	Timeout      int      `json:"timeout"`
	CacheTTL     int      `json:"cache_ttl"`
}

// MetricsResponse represents a metrics response
type MetricsResponse struct {
	RequestCount     map[string]int64   `json:"request_count"`
	AverageLatency   map[string]float64 `json:"average_latency"`
	ErrorRate        map[string]float64 `json:"error_rate"`
	StatusCodeCounts map[int]int64      `json:"status_code_counts"`
	ServiceLatency   map[string]float64 `json:"service_latency"`
}
```

### HTTP Server

The HTTP server handles HTTP requests and responses.

#### `server.go`

```go
package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server represents an HTTP server
type Server struct {
	server *http.Server
	logger Logger
}

// NewServer creates a new Server
func NewServer(handler http.Handler, port int, readTimeout, writeTimeout, shutdownTimeout time.Duration, logger Logger) *Server {
	return &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  120 * time.Second,
		},
		logger: logger,
	}
}

// Start starts the server
func (s *Server) Start() error {
	// Start server in a goroutine
	go func() {
		s.logger.Info("Starting server", "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down server")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown error", "error", err)
		return err
	}

	s.logger.Info("Server stopped")
	return nil
}
```

## Conclusion

The interfaces layer provides the entry point to the application and handles the communication between the outside world and the application. It includes HTTP handlers, middleware, and router implementations that connect the application to the outside world.

Key components of the interfaces layer include:

1. **API Handlers**: Handle HTTP requests and responses.
2. **Middleware**: Handle cross-cutting concerns like logging, authentication, and rate limiting.
3. **Router**: Handles request routing.
4. **Interface DTOs**: Define the data transfer objects for the interfaces layer.
5. **HTTP Server**: Handles HTTP requests and responses.

The interfaces layer depends on the application layer and provides the concrete implementations that allow the application to communicate with the outside world.
