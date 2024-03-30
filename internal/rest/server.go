package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "github.com/marco-almeida/gobank/configs"
	"github.com/marco-almeida/gobank/internal/handler"
	"github.com/marco-almeida/gobank/internal/postgres"
	"github.com/marco-almeida/gobank/internal/service"
	"github.com/sirupsen/logrus"
)

type APIServer struct {
	addr string
	log  *logrus.Logger
}

func NewAPIServer(addr string, logger *logrus.Logger) *APIServer {
	return &APIServer{
		addr: addr,
		log:  logger,
	}
}

func (s *APIServer) Serve() {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           http.NewServeMux(),
		ReadTimeout:       time.Hour,
		WriteTimeout:      time.Hour,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       time.Hour,
	}
	// add /api prefix to all routes
	srv.Handler.(*http.ServeMux).Handle("/api/", http.StripPrefix("/api", srv.Handler))

	// userService := NewUserService(s.log, s.store)
	// userService.RegisterRoutes(srv.Handler.(*http.ServeMux))

	// accountsService := NewAccountsService(s.log, s.store)
	// accountsService.RegisterRoutes(srv.Handler.(*http.ServeMux))

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Envs.PgUser,
		config.Envs.PgPassword,
		config.Envs.PgHost,
		config.Envs.Port,
		config.Envs.PgDb)
	userPostgresStorage := postgres.NewUser(connStr, s.log)

	err := userPostgresStorage.Init()
	if err != nil {
		s.log.Fatal(err)
	}

	userService := service.NewUser(userPostgresStorage, s.log)
	handler.NewUser(userService, s.log).RegisterRoutes(srv.Handler.(*http.ServeMux))

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
