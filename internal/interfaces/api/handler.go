package api

import (
	"encoding/json"
	"net/http"

	"api-gateway-sample/internal/application/usecase"
	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/pkg/logger"
)

// Handler handles HTTP requests
type Handler struct {
	proxyUseCase             *usecase.ProxyUseCase
	authUseCase              *usecase.AuthUseCase
	rateLimitUseCase         *usecase.RateLimitUseCase
	serviceManagementUseCase *usecase.ServiceManagementUseCase
	logger                   logger.Logger
}

// NewHandler creates a new Handler instance
func NewHandler(
	proxyUseCase *usecase.ProxyUseCase,
	authUseCase *usecase.AuthUseCase,
	rateLimitUseCase *usecase.RateLimitUseCase,
	serviceManagementUseCase *usecase.ServiceManagementUseCase,
	logger logger.Logger,
) *Handler {
	return &Handler{
		proxyUseCase:             proxyUseCase,
		authUseCase:              authUseCase,
		rateLimitUseCase:         rateLimitUseCase,
		serviceManagementUseCase: serviceManagementUseCase,
		logger:                   logger,
	}
}

// ProxyHandler handles proxy requests
func (h *Handler) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Create request entity
	request := &entity.Request{
		ID:          r.Header.Get("X-Request-ID"),
		Method:      r.Method,
		Path:        r.URL.Path,
		Headers:     r.Header,
		QueryParams: r.URL.Query(),
		ClientIP:    r.RemoteAddr,
	}

	// Read request body if present
	if r.Body != nil {
		body, err := readBody(r)
		if err != nil {
			h.handleError(w, err, http.StatusBadRequest)
			return
		}
		request.Body = body
	}

	// Proxy request
	response, err := h.proxyUseCase.ProxyRequest(r.Context(), request)
	if err != nil {
		h.handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Write response
	h.writeResponse(w, response)
}

// HealthCheckHandler handles health check requests
func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Helper functions

func readBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return json.Marshal(r.Body)
}

func (h *Handler) handleError(w http.ResponseWriter, err error, statusCode int) {
	h.logger.Error("Request failed", "error", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func (h *Handler) writeResponse(w http.ResponseWriter, response *entity.Response) {
	// Set headers
	for key, values := range response.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code
	w.WriteHeader(response.StatusCode)

	// Write body
	w.Write(response.Body)
}
