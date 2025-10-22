package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/hoenirvili/axiogate/log"
)

type sig struct {
	ctx  context.Context
	stop context.CancelFunc
}

type Server struct {
	s   http.Server
	sig sig
	log *slog.Logger
}

type Option func(s *Server)

func WithWhenToClose(ctx context.Context, stop context.CancelFunc) Option {
	return func(s *Server) {
		s.sig = sig{ctx, stop}
	}
}

// WithLogger sets a custom logger.
func WithLogger(log *slog.Logger) Option {
	return func(s *Server) {
		s.log = log
	}
}

const port = 8080

// NewServer is an http server that serves the static files
// the spa app and has the rest api.
func NewServer(options ...Option) *Server {
	s := &Server{
		s: http.Server{
			Addr: fmt.Sprintf(":%d", port),
		},
		log: log.Noop(),
		sig: sig{
			ctx:  context.Background(),
			stop: func() {},
		},
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// Route interface defines how routes are set.
type Route interface {
	// Append appends the handler route in the root mux.
	Append(mux *http.ServeMux)
}

// Routes mount all routnes under the server.
func (s *Server) Routes(routes ...Route) {
	mux := http.NewServeMux()
	for _, route := range routes {
		route.Append(mux)
	}
	s.s.Handler = mux
}

func (s *Server) start() chan error {
	errs := make(chan error)

	go func() {
		defer close(errs)
		var err error
		s.log.With(
			slog.Int("port", port),
			slog.String("addr", "127.0.0.1"),
		).Info("Http server start")
		if err = s.s.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			errs <- fmt.Errorf("failed to start server, %w", err)
		}
	}()

	return errs
}

// Start starts the http server with default configuration.
func Start(s *Server) error {
	errs := s.start()

	select {
	case <-s.sig.ctx.Done():
		s.log.Info("Shutdown received, closing gracefully")
	case err := <-errs:
		return err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := s.s.Shutdown(ctx); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to shutdown gracefully, %w", err)
	}

	return nil
}
