package entity

import (
	"time"
)

// Request represents a client request to the API Gateway
type Request struct {
	ID            string
	Method        string
	Path          string
	Headers       map[string][]string
	QueryParams   map[string][]string
	Body          []byte
	ClientIP      string
	Timestamp     time.Time
	Authenticated bool
	UserID        string
	Timeout       time.Duration
}

// NewRequest creates a new Request instance
func NewRequest(
	method string,
	path string,
	headers map[string][]string,
	queryParams map[string][]string,
	body []byte,
	clientIP string,
) *Request {
	return &Request{
		ID:            generateRequestID(),
		Method:        method,
		Path:          path,
		Headers:       headers,
		QueryParams:   queryParams,
		Body:          body,
		ClientIP:      clientIP,
		Timestamp:     time.Now(),
		Authenticated: false,
		Timeout:       30 * time.Second,
	}
}

// SetAuthenticated sets the authentication status and user ID
func (r *Request) SetAuthenticated(authenticated bool, userID string) {
	r.Authenticated = authenticated
	r.UserID = userID
}

// SetTimeout sets the request timeout
func (r *Request) SetTimeout(timeout time.Duration) {
	r.Timeout = timeout
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
