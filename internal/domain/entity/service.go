package entity

import (
	"fmt"
	"net/url"
	"strings"
)

// Service represents a backend service that can be accessed through the API Gateway
type Service struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	BaseURL     string            `json:"baseUrl"`
	Timeout     int               `json:"timeout"`
	RetryCount  int               `json:"retryCount"`
	IsActive    bool              `json:"isActive"`
	Metadata    map[string]string `json:"metadata"`
	Endpoints   []Endpoint        `json:"endpoints"`
}

// Endpoint represents a service endpoint configuration
type Endpoint struct {
	Path           string   `json:"path"`
	Methods        []string `json:"methods"`
	RateLimit      int      `json:"rateLimit"`
	AuthRequired   bool     `json:"authRequired"`
	Timeout        int      `json:"timeout"` // in seconds
	RetryCount     int      `json:"retryCount"`
	RetryDelay     int      `json:"retryDelay"` // in milliseconds
	CacheTTL       int      `json:"cacheTTL"`   // in seconds
	CircuitBreaker struct {
		Enabled          bool    `json:"enabled"`
		FailureThreshold float64 `json:"failureThreshold"`
		MinRequestCount  int     `json:"minRequestCount"`
		BreakDuration    int     `json:"breakDuration"` // in seconds
		HalfOpenRequests int     `json:"halfOpenRequests"`
	} `json:"circuitBreaker"`
	Cache struct {
		Enabled bool `json:"enabled"`
		TTL     int  `json:"ttl"` // in seconds
	} `json:"cache"`
	Transform struct {
		Request  map[string]string `json:"request"`  // header transformations
		Response map[string]string `json:"response"` // header transformations
	} `json:"transform"`
}

// NewService creates a new Service instance
func NewService(
	id string,
	name string,
	version string,
	description string,
	baseURL string,
	timeout int,
	retryCount int,
) *Service {
	return &Service{
		ID:          id,
		Name:        name,
		Version:     version,
		Description: description,
		BaseURL:     baseURL,
		Timeout:     timeout,
		RetryCount:  retryCount,
		IsActive:    true,
		Metadata:    make(map[string]string),
		Endpoints:   make([]Endpoint, 0),
	}
}

// AddEndpoint adds a new endpoint to the service
func (s *Service) AddEndpoint(endpoint Endpoint) {
	s.Endpoints = append(s.Endpoints, endpoint)
}

// FindEndpoint finds an endpoint by path and method
func (s *Service) FindEndpoint(path string, method string) *Endpoint {
	for _, endpoint := range s.Endpoints {
		if endpoint.Path == path {
			for _, m := range endpoint.Methods {
				if m == method {
					return &endpoint
				}
			}
		}
	}
	return nil
}

// SetActive sets the service active status
func (s *Service) SetActive(active bool) {
	s.IsActive = active
}

// AddMetadata adds metadata to the service
func (s *Service) AddMetadata(key string, value string) {
	s.Metadata[key] = value
}

// Validate validates the service configuration
func (s *Service) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("service name is required")
	}

	if s.BaseURL == "" {
		return fmt.Errorf("service base URL is required")
	}

	if _, err := url.Parse(s.BaseURL); err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	if len(s.Endpoints) == 0 {
		return fmt.Errorf("at least one endpoint is required")
	}

	for i, endpoint := range s.Endpoints {
		if err := endpoint.Validate(); err != nil {
			return fmt.Errorf("invalid endpoint at index %d: %w", i, err)
		}
	}

	return nil
}

// Validate validates the endpoint configuration
func (e *Endpoint) Validate() error {
	if e.Path == "" {
		return fmt.Errorf("endpoint path is required")
	}

	if !strings.HasPrefix(e.Path, "/") {
		return fmt.Errorf("endpoint path must start with /")
	}

	if len(e.Methods) == 0 {
		return fmt.Errorf("at least one HTTP method is required")
	}

	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
	}

	for _, method := range e.Methods {
		if !validMethods[method] {
			return fmt.Errorf("invalid HTTP method: %s", method)
		}
	}

	if e.RateLimit < 0 {
		return fmt.Errorf("rate limit cannot be negative")
	}

	if e.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	if e.RetryCount < 0 {
		return fmt.Errorf("retry count cannot be negative")
	}

	if e.RetryDelay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}

	if e.CircuitBreaker.Enabled {
		if e.CircuitBreaker.FailureThreshold < 0 || e.CircuitBreaker.FailureThreshold > 1 {
			return fmt.Errorf("circuit breaker failure threshold must be between 0 and 1")
		}

		if e.CircuitBreaker.MinRequestCount < 0 {
			return fmt.Errorf("circuit breaker minimum request count cannot be negative")
		}

		if e.CircuitBreaker.BreakDuration < 0 {
			return fmt.Errorf("circuit breaker break duration cannot be negative")
		}

		if e.CircuitBreaker.HalfOpenRequests < 0 {
			return fmt.Errorf("circuit breaker half-open requests cannot be negative")
		}
	}

	if e.Cache.Enabled && e.Cache.TTL < 0 {
		return fmt.Errorf("cache TTL cannot be negative")
	}

	return nil
}
