package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockLogger is a simple mock implementation of the logger interface
type MockLogger struct {
	debugCalled bool
	infoCalled  bool
	warnCalled  bool
	errorCalled bool
	fatalCalled bool
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.debugCalled = true
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.infoCalled = true
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.warnCalled = true
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.errorCalled = true
}

func (m *MockLogger) Fatal(msg string, args ...interface{}) {
	m.fatalCalled = true
}

func TestLoggingMiddlewareSimple(t *testing.T) {
	// Create a mock logger
	mockLogger := &MockLogger{}

	// Create a router with the mock logger
	router := &Router{
		logger: mockLogger,
	}

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply the logging middleware
	handler := router.loggingMiddleware(testHandler)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())

	// Verify that the logger was called
	assert.True(t, mockLogger.infoCalled, "Info method should have been called")
}

func TestRecoveryMiddlewareSimple(t *testing.T) {
	// Create a mock logger
	mockLogger := &MockLogger{}

	// Create a router with the mock logger
	router := &Router{
		logger: mockLogger,
	}

	// Create a test handler that panics
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Apply the recovery middleware
	handler := router.recoveryMiddleware(testHandler)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Verify the response
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	// Verify that the logger was called
	assert.True(t, mockLogger.errorCalled, "Error method should have been called")
}

func TestCorsMiddlewareSimple(t *testing.T) {
	// Create a router
	router := &Router{}

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply the CORS middleware
	handler := router.corsMiddleware(testHandler)

	// Test cases
	testCases := []struct {
		name           string
		method         string
		origin         string
		expectedStatus int
		expectedHeader string
	}{
		{
			name:           "Normal request",
			method:         http.MethodGet,
			origin:         "http://example.com",
			expectedStatus: http.StatusOK,
			expectedHeader: "*",
		},
		{
			name:           "Preflight request",
			method:         http.MethodOptions,
			origin:         "http://example.com",
			expectedStatus: http.StatusOK,
			expectedHeader: "*",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(tc.method, "/test", nil)
			req.Header.Set("Origin", tc.origin)

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rr, req)

			// Verify the response
			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Equal(t, tc.expectedHeader, rr.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}
