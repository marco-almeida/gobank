package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	config "github.com/marco-almeida/gobank/configs"
	"github.com/marco-almeida/gobank/internal/rest"
	"github.com/marco-almeida/gobank/internal/storage"
	"github.com/marco-almeida/gobank/pkg/logger"
)

func main() {
	listenAddr := flag.String("listen-addr", ":3000", "server listen address")
	flag.Parse()

	// set up logging
	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		panic(err)
	}
	logFile := filepath.Join("logs", "main.log")
	logger := logger.New(logFile, true)

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Envs.PgUser,
		config.Envs.PgPassword,
		config.Envs.PgHost,
		config.Envs.Port,
		config.Envs.PgDb)
	storage := storage.NewPostgresStorage(connStr, logger)

	err = storage.Init()
	if err != nil {
		logger.Fatal(err)
	}

	server := rest.NewAPIServer(*listenAddr, logger, storage)
	server.Serve()
}
