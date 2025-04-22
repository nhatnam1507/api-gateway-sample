package entity

import (
	"testing"
	"time"
)

func TestNewRequest(t *testing.T) {
	// Test data
	method := "GET"
	path := "/api/test"
	headers := map[string][]string{
		"Content-Type": {"application/json"},
	}
	queryParams := map[string][]string{
		"param1": {"value1"},
	}
	body := []byte(`{"test":"data"}`)
	clientIP := "127.0.0.1"

	// Create new request
	request := NewRequest(method, path, headers, queryParams, body, clientIP)

	// Verify request properties
	if request.ID == "" {
		t.Error("Request ID should not be empty")
	}
	if request.Method != method {
		t.Errorf("Expected method %s, got %s", method, request.Method)
	}
	if request.Path != path {
		t.Errorf("Expected path %s, got %s", path, request.Path)
	}
	if len(request.Headers) != len(headers) {
		t.Errorf("Expected %d headers, got %d", len(headers), len(request.Headers))
	}
	if len(request.QueryParams) != len(queryParams) {
		t.Errorf("Expected %d query params, got %d", len(queryParams), len(request.QueryParams))
	}
	if string(request.Body) != string(body) {
		t.Errorf("Expected body %s, got %s", string(body), string(request.Body))
	}
	if request.ClientIP != clientIP {
		t.Errorf("Expected client IP %s, got %s", clientIP, request.ClientIP)
	}
	if request.Authenticated {
		t.Error("Request should not be authenticated by default")
	}
	if request.UserID != "" {
		t.Error("User ID should be empty by default")
	}
	if request.Timeout != 30*time.Second {
		t.Errorf("Expected timeout %v, got %v", 30*time.Second, request.Timeout)
	}
}

func TestRequest_SetAuthenticated(t *testing.T) {
	// Create new request
	request := NewRequest("GET", "/api/test", nil, nil, nil, "127.0.0.1")

	// Set authenticated
	userID := "user123"
	request.SetAuthenticated(true, userID)

	// Verify authentication status
	if !request.Authenticated {
		t.Error("Request should be authenticated")
	}
	if request.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, request.UserID)
	}

	// Set unauthenticated
	request.SetAuthenticated(false, "")

	// Verify authentication status
	if request.Authenticated {
		t.Error("Request should not be authenticated")
	}
	if request.UserID != "" {
		t.Errorf("Expected empty user ID, got %s", request.UserID)
	}
}

func TestRequest_SetTimeout(t *testing.T) {
	// Create new request
	request := NewRequest("GET", "/api/test", nil, nil, nil, "127.0.0.1")

	// Set timeout
	timeout := 60 * time.Second
	request.SetTimeout(timeout)

	// Verify timeout
	if request.Timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, request.Timeout)
	}
}

func TestGenerateRequestID(t *testing.T) {
	// Generate multiple request IDs
	id1 := generateRequestID()
	// Add a small delay to ensure different timestamps
	time.Sleep(10 * time.Millisecond)
	id2 := generateRequestID()

	// Verify IDs are not empty
	if id1 == "" {
		t.Error("Generated request ID should not be empty")
	}
	if id2 == "" {
		t.Error("Generated request ID should not be empty")
	}

	// Verify IDs are unique
	if id1 == id2 {
		t.Error("Generated request IDs should be unique")
	}
}
