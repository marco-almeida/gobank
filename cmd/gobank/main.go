package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/marco-almeida/gobank/cmd/internal"
	"github.com/marco-almeida/gobank/internal/handler"
	"github.com/marco-almeida/gobank/internal/middleware"
	postgres "github.com/marco-almeida/gobank/internal/postgresql"
	"github.com/marco-almeida/gobank/internal/service"
	"github.com/marco-almeida/gobank/pkg/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	errC, err := run(&internal.Envs)
	if err != nil {
		log.Fatalf("Couldn't run: %s", err)
	}

	if err := <-errC; err != nil {
		log.Fatalf("Error while running: %s", err)
	}
}

func run(cfg *internal.Config) (<-chan error, error) {
	// set up logging
	logFolder := filepath.Join("logs", "gobank")
	err := os.MkdirAll(logFolder, os.ModePerm)
	if err != nil {
		return nil, err
	}

	logFile := filepath.Join(logFolder, "main.log")
	logger := logger.New(logFile, true)

	//

	db, err := internal.NewPostgreSQL(cfg)
	if err != nil {
		return nil, err
	}

	srv, err := newServer(serverConfig{
		Address: cfg.GobankAddress,
		DB:      db,
		Logger:  logger,
		Envs:    cfg,
	})

	if err != nil {
		return nil, err
	}

	errC := make(chan error, 1)

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-ctx.Done()

		logger.Info("Shutdown signal received")

		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer func() {
			db.Close()
			stop()
			cancel()
			close(errC)
		}()

		srv.SetKeepAlivesEnabled(false)

		if err := srv.Shutdown(ctxTimeout); err != nil { //nolint: contextcheck
			errC <- err
		}

		logger.Info("Shutdown completed")
	}()

	go func() {
		logger.Infof("Listening and serving %s", cfg.GobankAddress)

		// "ListenAndServe always returns a non-nil error. After Shutdown or Close, the returned error is
		// ErrServerClosed."
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errC <- err
		}
	}()

	return errC, nil
}

type serverConfig struct {
	Address string
	DB      *sql.DB
	Logger  *logrus.Logger
	Envs    *internal.Config
}

func newServer(conf serverConfig) (*http.Server, error) {
	srv := &http.Server{
		Addr:              conf.Address,
		Handler:           http.NewServeMux(),
		ReadTimeout:       time.Hour,
		WriteTimeout:      time.Hour,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       time.Hour,
	}

	// add /api prefix to all routes
	srv.Handler.(*http.ServeMux).Handle("/api/", http.StripPrefix("/api", srv.Handler))
	// Users service
	userRepo := postgres.NewUser(conf.DB)
	err := userRepo.Init()
	if err != nil {
		return nil, err
	}

	userService := service.NewUser(userRepo, conf.Logger)
	handler.NewUser(userService, conf.Logger).RegisterRoutes(srv.Handler.(*http.ServeMux))

	service.InitAuth(conf.Envs.JWTSecret)
	// Middleware
	loggingMiddleware := middleware.LoggingMiddleware(conf.Logger)
	srv.Handler = loggingMiddleware(middleware.RateLimiterMiddleware(srv.Handler))

	return srv, nil
}
