package main

import (
	"os"
	"path/filepath"

	"github.com/marco-almeida/gobank/internal/rest"
	"github.com/marco-almeida/gobank/pkg/logger"
)

func main() {
	// set up logging
	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		panic(err)
	}
	logFile := filepath.Join("logs", "main.log")
	l := logger.New(logFile, true)

	server := rest.NewAPIServer(":3000", l)
	server.Serve()
}
