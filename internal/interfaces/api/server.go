package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway-sample/pkg/logger"
)

// Server represents the HTTP server
type Server struct {
	server   *http.Server
	logger   logger.Logger
	shutdown chan os.Signal
}

// NewServer creates a new Server instance
func NewServer(handler http.Handler, port int, readTimeout, writeTimeout, shutdownTimeout time.Duration, logger logger.Logger) *Server {
	return &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
		logger:   logger,
		shutdown: make(chan os.Signal, 1),
	}
}

// Start starts the server
func (s *Server) Start() error {
	// Set up signal handling
	signal.Notify(s.shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		s.logger.Info("Starting server", "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server failed", "error", err)
		}
	}()

	// Wait for shutdown signal
	<-s.shutdown
	s.logger.Info("Server shutting down")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown failed", "error", err)
		return err
	}

	s.logger.Info("Server stopped gracefully")
	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	s.shutdown <- syscall.SIGTERM
	return nil
}
