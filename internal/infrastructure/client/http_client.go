package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/pkg/logger"
)

// HTTPClient implements an HTTP client for communicating with backend services
type HTTPClient struct {
	client *http.Client
	logger logger.Logger
}

// NewHTTPClient creates a new HTTPClient instance
func NewHTTPClient(timeout time.Duration, logger logger.Logger) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger: logger,
	}
}

// SendRequest sends an HTTP request to a backend service
func (c *HTTPClient) SendRequest(ctx context.Context, request *entity.Request, service *entity.Service) (*entity.Response, error) {
	startTime := time.Now()

	// Create target URL
	targetURL := fmt.Sprintf("%s%s", service.BaseURL, request.Path)
	if request.QueryParams != nil && len(request.QueryParams) > 0 {
		targetURL += "?"
		for key, values := range request.QueryParams {
			for _, value := range values {
				targetURL += fmt.Sprintf("%s=%s&", key, value)
			}
		}
		targetURL = targetURL[:len(targetURL)-1] // Remove trailing &
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, request.Method, targetURL, bytes.NewReader(request.Body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Copy headers
	for key, values := range request.Headers {
		for _, value := range values {
			httpReq.Header.Add(key, value)
		}
	}

	// Add X-Forwarded headers
	httpReq.Header.Set("X-Forwarded-For", request.ClientIP)
	httpReq.Header.Set("X-Request-ID", request.ID)

	// Send request
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Create response
	response := &entity.Response{
		RequestID:    request.ID,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		Body:         body,
		ContentType:  httpResp.Header.Get("Content-Type"),
		Timestamp:    time.Now(),
		LatencyMs:    time.Since(startTime).Milliseconds(),
		CachedResult: false,
	}

	// Log request details
	c.logger.Info("Request completed",
		"request_id", request.ID,
		"method", request.Method,
		"path", request.Path,
		"service", service.Name,
		"status", response.StatusCode,
		"latency_ms", response.LatencyMs,
	)

	return response, nil
}
