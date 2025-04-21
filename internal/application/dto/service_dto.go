package dto

import (
	"api-gateway-sample/internal/domain/entity"
)

// CreateServiceRequest represents a request to create a new service
type CreateServiceRequest struct {
	Name      string           `json:"name" validate:"required"`
	BaseURL   string           `json:"baseUrl" validate:"required,url"`
	Endpoints []EndpointConfig `json:"endpoints" validate:"required,dive"`
}

// EndpointConfig represents the configuration for a service endpoint
type EndpointConfig struct {
	Path           string   `json:"path" validate:"required"`
	Methods        []string `json:"methods" validate:"required,dive,oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
	RateLimit      int      `json:"rateLimit" validate:"min=0"`
	AuthRequired   bool     `json:"authRequired"`
	Timeout        int      `json:"timeout" validate:"min=0"` // in seconds
	RetryCount     int      `json:"retryCount" validate:"min=0"`
	RetryDelay     int      `json:"retryDelay" validate:"min=0"` // in milliseconds
	CircuitBreaker struct {
		Enabled          bool    `json:"enabled"`
		FailureThreshold float64 `json:"failureThreshold" validate:"min=0,max=1"`
		MinRequestCount  int     `json:"minRequestCount" validate:"min=0"`
		BreakDuration    int     `json:"breakDuration" validate:"min=0"` // in seconds
		HalfOpenRequests int     `json:"halfOpenRequests" validate:"min=0"`
	} `json:"circuitBreaker"`
	Cache struct {
		Enabled bool `json:"enabled"`
		TTL     int  `json:"ttl" validate:"min=0"` // in seconds
	} `json:"cache"`
	Transform struct {
		Request  map[string]string `json:"request"`  // header transformations
		Response map[string]string `json:"response"` // header transformations
	} `json:"transform"`
}

// UpdateServiceRequest represents a request to update an existing service
type UpdateServiceRequest struct {
	Name      string           `json:"name" validate:"required"`
	BaseURL   string           `json:"baseUrl" validate:"required,url"`
	Endpoints []EndpointConfig `json:"endpoints" validate:"required,dive"`
}

// ServiceResponse represents a service in API responses
type ServiceResponse struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	BaseURL   string           `json:"baseUrl"`
	Endpoints []EndpointConfig `json:"endpoints"`
}

// ToEntity converts a CreateServiceRequest to a Service entity
func (r *CreateServiceRequest) ToEntity() *entity.Service {
	endpoints := make([]entity.Endpoint, len(r.Endpoints))
	for i, e := range r.Endpoints {
		endpoints[i] = entity.Endpoint{
			Path:         e.Path,
			Methods:      e.Methods,
			RateLimit:    e.RateLimit,
			AuthRequired: e.AuthRequired,
			Timeout:      e.Timeout,
			RetryCount:   e.RetryCount,
			RetryDelay:   e.RetryDelay,
			CircuitBreaker: struct {
				Enabled          bool    `json:"enabled"`
				FailureThreshold float64 `json:"failureThreshold"`
				MinRequestCount  int     `json:"minRequestCount"`
				BreakDuration    int     `json:"breakDuration"`
				HalfOpenRequests int     `json:"halfOpenRequests"`
			}{
				Enabled:          e.CircuitBreaker.Enabled,
				FailureThreshold: e.CircuitBreaker.FailureThreshold,
				MinRequestCount:  e.CircuitBreaker.MinRequestCount,
				BreakDuration:    e.CircuitBreaker.BreakDuration,
				HalfOpenRequests: e.CircuitBreaker.HalfOpenRequests,
			},
			Cache: struct {
				Enabled bool `json:"enabled"`
				TTL     int  `json:"ttl"`
			}{
				Enabled: e.Cache.Enabled,
				TTL:     e.Cache.TTL,
			},
			Transform: struct {
				Request  map[string]string `json:"request"`
				Response map[string]string `json:"response"`
			}{
				Request:  e.Transform.Request,
				Response: e.Transform.Response,
			},
		}
	}

	return &entity.Service{
		Name:      r.Name,
		BaseURL:   r.BaseURL,
		Endpoints: endpoints,
	}
}

// FromEntity creates a ServiceResponse from a Service entity
func FromEntity(s *entity.Service) *ServiceResponse {
	endpoints := make([]EndpointConfig, len(s.Endpoints))
	for i, e := range s.Endpoints {
		endpoints[i] = EndpointConfig{
			Path:         e.Path,
			Methods:      e.Methods,
			RateLimit:    e.RateLimit,
			AuthRequired: e.AuthRequired,
			Timeout:      e.Timeout,
			RetryCount:   e.RetryCount,
			RetryDelay:   e.RetryDelay,
			CircuitBreaker: struct {
				Enabled          bool    `json:"enabled"`
				FailureThreshold float64 `json:"failureThreshold" validate:"min=0,max=1"`
				MinRequestCount  int     `json:"minRequestCount" validate:"min=0"`
				BreakDuration    int     `json:"breakDuration" validate:"min=0"`
				HalfOpenRequests int     `json:"halfOpenRequests" validate:"min=0"`
			}{
				Enabled:          e.CircuitBreaker.Enabled,
				FailureThreshold: e.CircuitBreaker.FailureThreshold,
				MinRequestCount:  e.CircuitBreaker.MinRequestCount,
				BreakDuration:    e.CircuitBreaker.BreakDuration,
				HalfOpenRequests: e.CircuitBreaker.HalfOpenRequests,
			},
			Cache: struct {
				Enabled bool `json:"enabled"`
				TTL     int  `json:"ttl" validate:"min=0"`
			}{
				Enabled: e.Cache.Enabled,
				TTL:     e.Cache.TTL,
			},
			Transform: struct {
				Request  map[string]string `json:"request"`
				Response map[string]string `json:"response"`
			}{
				Request:  e.Transform.Request,
				Response: e.Transform.Response,
			},
		}
	}

	return &ServiceResponse{
		ID:        s.ID,
		Name:      s.Name,
		BaseURL:   s.BaseURL,
		Endpoints: endpoints,
	}
}
