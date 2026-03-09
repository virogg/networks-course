package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/virogg/networks-course/service/pkg/logger"
)

type ServerConfig struct {
	Port            string
	Handler         http.Handler
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Server struct {
	httpServer      *http.Server
	log             logger.Logger
	port            string
	shutdownTimeout time.Duration
}

func NewServer(log logger.Logger, cfg ServerConfig) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      cfg.Handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		log:             log,
		port:            cfg.Port,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.log.Info("starting http server", logger.NewField("port", s.port))
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		s.log.Info("shutting down server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown error: %w", err)
		}

		s.log.Info("server stopped gracefully")
	}

	return nil
}
