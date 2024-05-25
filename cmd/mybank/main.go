package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/marco-almeida/mybank/internal/config"
	"github.com/marco-almeida/mybank/internal/postgresql"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	// get env vars
	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	fmt.Printf("%+v\n", config)

	// config logging
	logFolder := filepath.Join("logs", "mybank")
	err = os.MkdirAll(logFolder, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create log folder")
	}

	logFile := filepath.Join(logFolder, "main.log")

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot open log file")
	}

	// log to file and terminal
	// set up human readable logging
	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: io.MultiWriter(os.Stdout, f)})
	} else {
		// set up json logging
		log.Logger = log.Output(io.MultiWriter(os.Stdout, f))
	}

	// init graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	// init db
	connPool, err := postgresql.NewPostgreSQL(ctx, &config)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	// run migrations
	dbSource := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", config.PostgresUser, config.PostgresPassword, config.PostgresHost, config.PostgresPort, config.PostgresDatabase)

	err = runDBMigration(config.MigrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot run db migration")
	}

	log.Info().Msg("db migrated successfully")

	// init server dependencies
	store := db.NewStore(connPool)

	waitGroup, ctx := errgroup.WithContext(ctx)

	// running in waitgroup coroutine in order to wait for graceful shutdown
	runHTPPServer(ctx, waitGroup, config, store)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error from wait group")
	}
}

func runHTPPServer(ctx context.Context, waitGroup *errgroup.Group, config config.Config, store db.Store) {
	server, err := newServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	waitGroup.Go(func() error {
		log.Info().Msg(fmt.Sprintf("start HTTP server on %s", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("cannot start server: %w", err)
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("shutting down gracefully, press Ctrl+C again to force")

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}

		log.Info().Msg("HTTP server stopped")

		return nil
	})
}

func runDBMigration(migrationURL string, dbSource string) error {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		return fmt.Errorf("cannot create new migrate instance: %w", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrate up: %w", err)
	}

	return nil
}

func newServer(config config.Config, store db.Store) (*http.Server, error) {
	router := gin.Default()

	if config.Environment != "development" && config.Environment != "testing" {
		gin.SetMode(gin.ReleaseMode)
	}

	srv := &http.Server{
		Addr:              config.HTTPServerAddress,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       10 * time.Second,
	}

	// TODO: add api prefix to all routes

	// init user repo

	// init user service

	// init user handler and register routes

	return srv, nil
}
