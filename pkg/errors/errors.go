package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Common errors
var (
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInternalServer     = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrTimeout            = errors.New("timeout")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrServiceNotFound    = errors.New("service not found")
)

// Error represents a custom error with additional context
type Error struct {
	Code    int
	Message string
	Err     error
}

// Error returns the error message
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Err
}

// NewError creates a new Error instance
func NewError(code int, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Is reports whether target matches the error
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return errors.Is(e.Err, target)
	}
	return e.Code == t.Code
}

// Common error codes
const (
	CodeNotFound           = 404
	CodeAlreadyExists      = 409
	CodeInvalidInput       = 400
	CodeUnauthorized       = 401
	CodeForbidden          = 403
	CodeInternalServer     = 500
	CodeServiceUnavailable = 503
	CodeTimeout            = 504
	CodeRateLimitExceeded  = 429
)

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// IsNotFound returns true if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists returns true if the error is an already exists error
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsInvalidInput returns true if the error is an invalid input error
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorized returns true if the error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden returns true if the error is a forbidden error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsInternalServer returns true if the error is an internal server error
func IsInternalServer(err error) bool {
	return errors.Is(err, ErrInternalServer)
}

// IsServiceUnavailable returns true if the error is a service unavailable error
func IsServiceUnavailable(err error) bool {
	return errors.Is(err, ErrServiceUnavailable)
}

// IsTimeout returns true if the error is a timeout error
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsRateLimitExceeded returns true if the error is a rate limit exceeded error
func IsRateLimitExceeded(err error) bool {
	return errors.Is(err, ErrRateLimitExceeded)
}

// Error represents an API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// New creates a new Error instance
func New(code int, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// WithDetails adds details to the error
func (e *APIError) WithDetails(details string) *APIError {
	e.Details = details
	return e
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// StatusCode returns the HTTP status code
func (e *APIError) StatusCode() int {
	return e.Code
}

// Common API errors
var (
	ErrBadRequest       = New(http.StatusBadRequest, "Bad request")
	ErrMethodNotAllowed = New(http.StatusMethodNotAllowed, "Method not allowed")
	ErrTooManyRequests  = New(http.StatusTooManyRequests, "Too many requests")
)

// IsAPIError checks if an error is an API error
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// ToAPIError converts an error to an API error
func ToAPIError(err error) *APIError {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return New(500, err.Error())
}
