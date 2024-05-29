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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/marco-almeida/mybank/internal/config"
	"github.com/marco-almeida/mybank/internal/handler"
	"github.com/marco-almeida/mybank/internal/postgresql"
	"github.com/marco-almeida/mybank/internal/service"
	"github.com/marco-almeida/mybank/internal/token"
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

	setupLogging(config)

	// setup graceful shutdown signals
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

	// running in waitgroup coroutine in order to wait for graceful shutdown
	waitGroup, ctx := errgroup.WithContext(ctx)
	runHTPPServer(ctx, waitGroup, config, connPool)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error from wait group")
	}
}

func setupLogging(config config.Config) {
	// log to file ./logs/mybank/main.log and terminal
	logFolder := filepath.Join("logs", "mybank")
	err := os.MkdirAll(logFolder, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create log folder")
	}

	logFile := filepath.Join(logFolder, "main.log")

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot open log file")
	}

	// set up json or human readable logging
	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: io.MultiWriter(os.Stdout, f)})
	} else {
		log.Logger = log.Output(io.MultiWriter(os.Stdout, f))
	}
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

func newServer(config config.Config, connPool *pgxpool.Pool) (*http.Server, error) {
	if config.Environment != "development" && config.Environment != "testing" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	srv := &http.Server{
		Addr:              config.HTTPServerAddress,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       10 * time.Second,
	}

	// init user repo
	userRepo := postgresql.NewUserRepository(connPool)

	// init session repo
	sessionRepo := postgresql.NewSessionRepository(connPool)

	// init user service
	userService := service.NewUserService(userRepo)

	// init token maker
	tokenMaker, err := token.NewJWTMaker(config.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	// init auth service
	authService := service.NewAuthService(*userService, sessionRepo, tokenMaker, config.AccessTokenDuration, config.RefreshTokenDuration)

	// init user handler and register routes
	handler.NewUserHandler(userService, authService).RegisterRoutes(router)

	// init account repo
	accountRepo := postgresql.NewAccountRepository(connPool)

	// init account service
	accountService := service.NewAccountService(accountRepo)

	// init account handler and register routes
	handler.NewAccountHandler(accountService).RegisterRoutes(router, tokenMaker)

	// init transfer repo
	transferRepo := postgresql.NewTransferRepository(connPool)

	// init transfer service
	transferService := service.NewTransferService(transferRepo)

	// init transfer handler and register routes
	handler.NewTransferHandler(transferService, accountService).RegisterRoutes(router, tokenMaker)

	return srv, nil
}

func runHTPPServer(ctx context.Context, waitGroup *errgroup.Group, config config.Config, connPool *pgxpool.Pool) {
	server, err := newServer(config, connPool)
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
