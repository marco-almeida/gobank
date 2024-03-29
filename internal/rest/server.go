package rest

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marco-almeida/gobank/internal/storage"
	"github.com/sirupsen/logrus"
)

type APIServer struct {
	addr  string
	log   *logrus.Logger
	store storage.Storer
}

func NewAPIServer(addr string, logger *logrus.Logger, store storage.Storer) *APIServer {
	return &APIServer{
		addr:  addr,
		log:   logger,
		store: store,
	}
}

func (s *APIServer) Serve() {
	srv := &http.Server{
		Addr:         s.addr,
		Handler:      http.NewServeMux(),
		ReadTimeout:  time.Hour,
		WriteTimeout: time.Hour,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:  time.Hour,
	}

	userService := NewUserService(s.log, s.store)
	userService.RegisterRoutes(srv.Handler.(*http.ServeMux))

	accountsService := NewAccountsService(s.log, s.store)
	accountsService.RegisterRoutes(srv.Handler.(*http.ServeMux))

	s.log.Info("Starting the API server at", s.addr)
	loggingMiddleware := LoggingMiddleware(s.log)
	srv.Handler = loggingMiddleware(RateLimiterMiddleware(srv.Handler))

	// graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.log.Fatalf("HTTP server error: %v", err)
		}
		s.log.Warn("Stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		s.log.Fatalf("HTTP shutdown error: %v", err)
	}
	s.log.Info("Graceful shutdown complete.")
}
