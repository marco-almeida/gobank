package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/marco-almeida/golang-api-project-layout/internal/rest"
	"github.com/marco-almeida/golang-api-project-layout/pkg/logger"
)

func main() {
	listenAddr := flag.String("listen-addr", ":3000", "server listen address")
	flag.Parse()
	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		panic(err)
	}
	logFile := filepath.Join("logs", "main.log")
	logger := logger.New(logFile, true)

	server := rest.NewAPIServer(*listenAddr, logger)
	server.Serve()
}
