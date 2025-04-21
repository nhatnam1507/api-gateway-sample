package entity

import (
	"testing"
)

func TestService_Validate(t *testing.T) {
	tests := []struct {
		name    string
		service *Service
		wantErr bool
	}{
		{
			name: "valid service",
			service: &Service{
				ID:      "1",
				Name:    "test-service",
				BaseURL: "http://localhost:8080",
				Endpoints: []Endpoint{
					{
						Path:         "/api/test",
						Methods:      []string{"GET", "POST"},
						RateLimit:    100,
						AuthRequired: true,
						Timeout:      30,
						RetryCount:   3,
						RetryDelay:   100,
						CircuitBreaker: struct {
							Enabled          bool    `json:"enabled"`
							FailureThreshold float64 `json:"failureThreshold"`
							MinRequestCount  int     `json:"minRequestCount"`
							BreakDuration    int     `json:"breakDuration"`
							HalfOpenRequests int     `json:"halfOpenRequests"`
						}{
							Enabled:          true,
							FailureThreshold: 0.5,
							MinRequestCount:  10,
							BreakDuration:    30,
							HalfOpenRequests: 5,
						},
						Cache: struct {
							Enabled bool `json:"enabled"`
							TTL     int  `json:"ttl"`
						}{
							Enabled: true,
							TTL:     300,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid service - empty name",
			service: &Service{
				ID:      "1",
				Name:    "",
				BaseURL: "http://localhost:8080",
				Endpoints: []Endpoint{
					{
						Path:    "/api/test",
						Methods: []string{"GET"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid service - empty base URL",
			service: &Service{
				ID:      "1",
				Name:    "test-service",
				BaseURL: "",
				Endpoints: []Endpoint{
					{
						Path:    "/api/test",
						Methods: []string{"GET"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid service - no endpoints",
			service: &Service{
				ID:        "1",
				Name:      "test-service",
				BaseURL:   "http://localhost:8080",
				Endpoints: []Endpoint{},
			},
			wantErr: true,
		},
		{
			name: "invalid service - invalid endpoint path",
			service: &Service{
				ID:      "1",
				Name:    "test-service",
				BaseURL: "http://localhost:8080",
				Endpoints: []Endpoint{
					{
						Path:    "",
						Methods: []string{"GET"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid service - invalid endpoint methods",
			service: &Service{
				ID:      "1",
				Name:    "test-service",
				BaseURL: "http://localhost:8080",
				Endpoints: []Endpoint{
					{
						Path:    "/api/test",
						Methods: []string{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid service - invalid rate limit",
			service: &Service{
				ID:      "1",
				Name:    "test-service",
				BaseURL: "http://localhost:8080",
				Endpoints: []Endpoint{
					{
						Path:      "/api/test",
						Methods:   []string{"GET"},
						RateLimit: -1,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid service - invalid timeout",
			service: &Service{
				ID:      "1",
				Name:    "test-service",
				BaseURL: "http://localhost:8080",
				Endpoints: []Endpoint{
					{
						Path:    "/api/test",
						Methods: []string{"GET"},
						Timeout: -1,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid service - invalid retry count",
			service: &Service{
				ID:      "1",
				Name:    "test-service",
				BaseURL: "http://localhost:8080",
				Endpoints: []Endpoint{
					{
						Path:       "/api/test",
						Methods:    []string{"GET"},
						RetryCount: -1,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid service - invalid circuit breaker threshold",
			service: &Service{
				ID:      "1",
				Name:    "test-service",
				BaseURL: "http://localhost:8080",
				Endpoints: []Endpoint{
					{
						Path:    "/api/test",
						Methods: []string{"GET"},
						CircuitBreaker: struct {
							Enabled          bool    `json:"enabled"`
							FailureThreshold float64 `json:"failureThreshold"`
							MinRequestCount  int     `json:"minRequestCount"`
							BreakDuration    int     `json:"breakDuration"`
							HalfOpenRequests int     `json:"halfOpenRequests"`
						}{
							Enabled:          true,
							FailureThreshold: 1.5,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.service.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEndpoint_Validate(t *testing.T) {
	tests := []struct {
		name     string
		endpoint *Endpoint
		wantErr  bool
	}{
		{
			name: "valid endpoint",
			endpoint: &Endpoint{
				Path:         "/api/test",
				Methods:      []string{"GET", "POST"},
				RateLimit:    100,
				AuthRequired: true,
				Timeout:      30,
				RetryCount:   3,
				RetryDelay:   100,
				CircuitBreaker: struct {
					Enabled          bool    `json:"enabled"`
					FailureThreshold float64 `json:"failureThreshold"`
					MinRequestCount  int     `json:"minRequestCount"`
					BreakDuration    int     `json:"breakDuration"`
					HalfOpenRequests int     `json:"halfOpenRequests"`
				}{
					Enabled:          true,
					FailureThreshold: 0.5,
					MinRequestCount:  10,
					BreakDuration:    30,
					HalfOpenRequests: 5,
				},
				Cache: struct {
					Enabled bool `json:"enabled"`
					TTL     int  `json:"ttl"`
				}{
					Enabled: true,
					TTL:     300,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid endpoint - empty path",
			endpoint: &Endpoint{
				Path:    "",
				Methods: []string{"GET"},
			},
			wantErr: true,
		},
		{
			name: "invalid endpoint - empty methods",
			endpoint: &Endpoint{
				Path:    "/api/test",
				Methods: []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid endpoint - invalid method",
			endpoint: &Endpoint{
				Path:    "/api/test",
				Methods: []string{"INVALID"},
			},
			wantErr: true,
		},
		{
			name: "invalid endpoint - negative rate limit",
			endpoint: &Endpoint{
				Path:      "/api/test",
				Methods:   []string{"GET"},
				RateLimit: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid endpoint - negative timeout",
			endpoint: &Endpoint{
				Path:    "/api/test",
				Methods: []string{"GET"},
				Timeout: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid endpoint - negative retry count",
			endpoint: &Endpoint{
				Path:       "/api/test",
				Methods:    []string{"GET"},
				RetryCount: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid endpoint - invalid circuit breaker threshold",
			endpoint: &Endpoint{
				Path:    "/api/test",
				Methods: []string{"GET"},
				CircuitBreaker: struct {
					Enabled          bool    `json:"enabled"`
					FailureThreshold float64 `json:"failureThreshold"`
					MinRequestCount  int     `json:"minRequestCount"`
					BreakDuration    int     `json:"breakDuration"`
					HalfOpenRequests int     `json:"halfOpenRequests"`
				}{
					Enabled:          true,
					FailureThreshold: 1.5,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.endpoint.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Endpoint.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
