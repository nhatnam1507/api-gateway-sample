package api

import (
	"context"
	"net/http"
	"time"

	"api-gateway-sample/internal/application/usecase"
	"api-gateway-sample/pkg/logger"

	"github.com/gorilla/mux"
)

// Router handles HTTP routing
type Router struct {
	handler          *Handler
	logger           logger.Logger
	authUseCase      *usecase.AuthUseCase
	rateLimitUseCase *usecase.RateLimitUseCase
}

// NewRouter creates a new Router instance
func NewRouter(
	handler *Handler,
	logger logger.Logger,
	authUseCase *usecase.AuthUseCase,
	rateLimitUseCase *usecase.RateLimitUseCase,
) *Router {
	return &Router{
		handler:          handler,
		logger:           logger,
		authUseCase:      authUseCase,
		rateLimitUseCase: rateLimitUseCase,
	}
}

// Setup sets up the router
func (r *Router) Setup() http.Handler {
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(
		r.loggingMiddleware,
		r.recoveryMiddleware,
		r.corsMiddleware,
	)

	// Health check route
	router.HandleFunc("/health", r.handler.HealthCheckHandler).Methods(http.MethodGet)

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.Use(r.authMiddleware)

	// Proxy routes
	api.PathPrefix("/v1/").Handler(http.HandlerFunc(r.handler.ProxyHandler))

	return router
}

// Middleware functions

func (r *Router) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w}

		// Call next handler
		next.ServeHTTP(rw, req)

		// Log request details
		r.logger.Info("Request completed",
			"method", req.Method,
			"path", req.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", req.RemoteAddr,
		)
	})
}

func (r *Router) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				r.logger.Error("Panic recovered", "error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, req)
	})
}

func (r *Router) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (r *Router) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Skip authentication for health check
		if req.URL.Path == "/health" {
			next.ServeHTTP(w, req)
			return
		}

		// Get token from Authorization header
		token := req.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := r.authUseCase.ValidateToken(req.Context(), token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := req.Context()
		for key, value := range claims {
			ctx = context.WithValue(ctx, key, value)
		}

		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
