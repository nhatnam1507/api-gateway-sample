package entity

import (
	"time"
)

// Response represents a response from a backend service
type Response struct {
	RequestID     string
	StatusCode    int
	Headers       map[string][]string
	Body          []byte
	ContentType   string
	ContentLength int
	Timestamp     time.Time
	LatencyMs     int64
	CachedResult  bool
}

// NewResponse creates a new Response instance
func NewResponse(
	requestID string,
	statusCode int,
	headers map[string][]string,
	body []byte,
) *Response {
	contentType := "application/json"
	if ct, ok := headers["Content-Type"]; ok && len(ct) > 0 {
		contentType = ct[0]
	}

	return &Response{
		RequestID:     requestID,
		StatusCode:    statusCode,
		Headers:       headers,
		Body:          body,
		ContentType:   contentType,
		ContentLength: len(body),
		Timestamp:     time.Now(),
		CachedResult:  false,
	}
}

// SetLatency sets the response latency in milliseconds
func (r *Response) SetLatency(startTime time.Time) {
	r.LatencyMs = time.Since(startTime).Milliseconds()
}

// SetCached sets whether the response was served from cache
func (r *Response) SetCached(cached bool) {
	r.CachedResult = cached
}
