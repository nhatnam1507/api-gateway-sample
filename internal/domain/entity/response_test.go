package entity

import (
	"testing"
	"time"
)

func TestNewResponse(t *testing.T) {
	// Test data
	requestID := "req123"
	statusCode := 200
	headers := map[string][]string{
		"Content-Type": {"application/json"},
	}
	body := []byte(`{"message":"success"}`)

	// Create new response
	response := NewResponse(requestID, statusCode, headers, body)

	// Verify response properties
	if response.RequestID != requestID {
		t.Errorf("Expected request ID %s, got %s", requestID, response.RequestID)
	}
	if response.StatusCode != statusCode {
		t.Errorf("Expected status code %d, got %d", statusCode, response.StatusCode)
	}
	if len(response.Headers) != len(headers) {
		t.Errorf("Expected %d headers, got %d", len(headers), len(response.Headers))
	}
	if string(response.Body) != string(body) {
		t.Errorf("Expected body %s, got %s", string(body), string(response.Body))
	}
	if response.ContentType != "application/json" {
		t.Errorf("Expected content type %s, got %s", "application/json", response.ContentType)
	}
	if response.ContentLength != len(body) {
		t.Errorf("Expected content length %d, got %d", len(body), response.ContentLength)
	}
	if response.CachedResult {
		t.Error("Response should not be cached by default")
	}
}

func TestResponse_SetLatency(t *testing.T) {
	// Create new response
	response := NewResponse("req123", 200, nil, nil)

	// Set latency
	startTime := time.Now().Add(-100 * time.Millisecond)
	response.SetLatency(startTime)

	// Verify latency is set and is approximately correct
	if response.LatencyMs < 90 || response.LatencyMs > 110 {
		t.Errorf("Expected latency around 100ms, got %dms", response.LatencyMs)
	}
}

func TestResponse_SetCached(t *testing.T) {
	// Create new response
	response := NewResponse("req123", 200, nil, nil)

	// Verify default cache status
	if response.CachedResult {
		t.Error("Response should not be cached by default")
	}

	// Set cached
	response.SetCached(true)

	// Verify cache status
	if !response.CachedResult {
		t.Error("Response should be cached")
	}

	// Set not cached
	response.SetCached(false)

	// Verify cache status
	if response.CachedResult {
		t.Error("Response should not be cached")
	}
}

func TestContentTypeExtraction(t *testing.T) {
	// Test with Content-Type header
	headers1 := map[string][]string{
		"Content-Type": {"text/plain"},
	}
	response1 := NewResponse("req1", 200, headers1, nil)
	if response1.ContentType != "text/plain" {
		t.Errorf("Expected content type %s, got %s", "text/plain", response1.ContentType)
	}

	// Test with no Content-Type header
	headers2 := map[string][]string{
		"X-Custom-Header": {"value"},
	}
	response2 := NewResponse("req2", 200, headers2, nil)
	if response2.ContentType != "application/json" {
		t.Errorf("Expected default content type %s, got %s", "application/json", response2.ContentType)
	}

	// Test with multiple Content-Type values
	headers3 := map[string][]string{
		"Content-Type": {"text/html", "application/xml"},
	}
	response3 := NewResponse("req3", 200, headers3, nil)
	if response3.ContentType != "text/html" {
		t.Errorf("Expected first content type %s, got %s", "text/html", response3.ContentType)
	}
}
