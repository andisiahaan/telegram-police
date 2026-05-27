package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const shutdownTimeout = 10 * time.Second

// Server wraps net/http.Server with graceful shutdown support.
type Server struct {
	httpServer *http.Server
}

// New creates a Server that listens on the given port and registers /webhook.
func New(port int, handler http.Handler) *Server {
	mux := http.NewServeMux()
	mux.Handle("/webhook", handler)

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Start runs the HTTP server in a blocking call.
// Returns a non-nil error only if the server stops for a reason other than Shutdown.
func (s *Server) Start() error {
	slog.Info("server listening", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the server, waiting up to shutdownTimeout for
// in-flight requests to finish.
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	slog.Info("shutting down server...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", "error", err)
	} else {
		slog.Info("server stopped cleanly")
	}
}
