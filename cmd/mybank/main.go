// package main

// import (
// 	"context"
// 	"database/sql"
// 	"errors"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"path/filepath"
// 	"syscall"
// 	"time"

// 	"github.com/marco-almeida/mybank/cmd/internal"
// 	"github.com/marco-almeida/mybank/internal/handler"
// 	"github.com/marco-almeida/mybank/internal/middleware"
// 	postgres "github.com/marco-almeida/mybank/internal/postgresql"
// 	"github.com/marco-almeida/mybank/internal/service"
// 	"github.com/marco-almeida/mybank/pkg/logger"
// 	"github.com/sirupsen/logrus"
// )

// func main() {
// 	errC, err := run(&internal.Envs)
// 	if err != nil {
// 		log.Fatalf("Couldn't run: %s", err)
// 	}

// 	if err := <-errC; err != nil {
// 		log.Fatalf("Error while running: %s", err)
// 	}
// }

// func run(cfg *internal.Config) (<-chan error, error) {
// 	// set up logging
// 	logFolder := filepath.Join("logs", "mybank")
// 	err := os.MkdirAll(logFolder, os.ModePerm)
// 	if err != nil {
// 		return nil, err
// 	}

// 	logFile := filepath.Join(logFolder, "main.log")
// 	logger := logger.New(logFile, true)

// 	//

// 	db, err := internal.NewPostgreSQL(cfg)
// 	if err != nil {
// 		return nil, err
// 	}

// 	srv, err := newServer(serverConfig{
// 		Address: cfg.mybankAddress,
// 		DB:      db,
// 		Logger:  logger,
// 		Envs:    cfg,
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	errC := make(chan error, 1)

// 	ctx, stop := signal.NotifyContext(context.Background(),
// 		os.Interrupt,
// 		syscall.SIGTERM,
// 		syscall.SIGQUIT)

// 	go func() {
// 		<-ctx.Done()

// 		logger.Info("Shutdown signal received")

// 		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)

// 		defer func() {
// 			db.Close()
// 			stop()
// 			cancel()
// 			close(errC)
// 		}()

// 		srv.SetKeepAlivesEnabled(false)

// 		if err := srv.Shutdown(ctxTimeout); err != nil { //nolint: contextcheck
// 			errC <- err
// 		}

// 		logger.Info("Shutdown completed")
// 	}()

// 	go func() {
// 		logger.Infof("Listening and serving %s", cfg.mybankAddress)

// 		// "ListenAndServe always returns a non-nil error. After Shutdown or Close, the returned error is
// 		// ErrServerClosed."
// 		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
// 			errC <- err
// 		}
// 	}()

// 	return errC, nil
// }

// type serverConfig struct {
// 	Address string
// 	DB      *sql.DB
// 	Logger  *logrus.Logger
// 	Envs    *internal.Config
// }

// func newServer(conf serverConfig) (*http.Server, error) {
// 	srv := &http.Server{
// 		Addr:              conf.Address,
// 		Handler:           http.NewServeMux(),
// 		ReadTimeout:       time.Hour,
// 		WriteTimeout:      time.Hour,
// 		ReadHeaderTimeout: 10 * time.Second,
// 		IdleTimeout:       time.Hour,
// 	}

// 	// add /api prefix to all routes
// 	srv.Handler.(*http.ServeMux).Handle("/api/", http.StripPrefix("/api", srv.Handler))

// 	//////////////////// Users service ////////////////////
// 	userRepo := postgres.NewUser(conf.DB)
// 	err := userRepo.Init()
// 	if err != nil {
// 		return nil, err
// 	}

// 	userService := service.NewUser(userRepo, conf.Logger)

// 	authService := service.NewAuth(conf.Logger, userService, conf.Envs.JWTSecret)
// 	handler.NewUser(userService, conf.Logger, authService).RegisterRoutes(srv.Handler.(*http.ServeMux))

// 	handler.NewAuth(authService, conf.Logger).RegisterRoutes(srv.Handler.(*http.ServeMux))

// 	// Accounts service
// 	accountRepo := postgres.NewAccount(conf.DB)
// 	err = accountRepo.Init()
// 	if err != nil {
// 		return nil, err
// 	}

// 	accountService := service.NewAccount(accountRepo, conf.Logger)
// 	handler.NewAccount(accountService, conf.Logger, authService).RegisterRoutes(srv.Handler.(*http.ServeMux))

// 	// Middleware
// 	loggingMiddleware := middleware.LoggingMiddleware(conf.Logger)
// 	srv.Handler = loggingMiddleware(middleware.RateLimiterMiddleware(srv.Handler))

// 	return srv, nil
// }

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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
	if config.Environment == "development" || config.Environment == "testing" {
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

	store := db.NewStore(connPool)

	waitGroup, ctx := errgroup.WithContext(ctx)

	runHTPPServer(ctx, waitGroup, config)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error from wait group")
	}
}

func runHTPPServer(ctx context.Context, waitGroup *errgroup.Group, config config.Config) {
	server, err := newServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
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
