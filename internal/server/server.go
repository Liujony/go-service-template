// Package server owns the HTTP lifecycle: listen, serve, drain on cancel.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Server wraps http.Server with a run/drain lifecycle that is testable.
type Server struct {
	srv     *http.Server
	log     *slog.Logger
	timeout time.Duration
	listen  net.Listener
}

// New builds the server. The caller owns the handler and the shutdown timeout.
func New(addr string, handler http.Handler, log *slog.Logger, timeout time.Duration) *Server {
	return &Server{
		srv: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		log:     log,
		timeout: timeout,
	}
}

// Addr returns the address the server is bound to. Only valid after Run started.
func (s *Server) Addr() string {
	if s.listen == nil {
		return ""
	}
	return s.listen.Addr().String()
}

// Run serves until ctx is cancelled, then drains in-flight requests for at most
// the configured timeout. It returns nil on a clean shutdown.
func (s *Server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.srv.Addr, err)
	}
	s.listen = listener
	s.log.Info("server listening", "addr", listener.Addr().String())

	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.log.Info("shutting down", "timeout", s.timeout.String())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.timeout)
		defer cancel()
		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return nil
	}
}
