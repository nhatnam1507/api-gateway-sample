package client

import "errors"

var (
	// ErrInvalidRequest is returned when the request is invalid
	ErrInvalidRequest = errors.New("invalid request")

	// ErrInvalidMethod is returned when the request method is invalid
	ErrInvalidMethod = errors.New("invalid method")

	// ErrInvalidPath is returned when the request path is invalid
	ErrInvalidPath = errors.New("invalid path")
)
