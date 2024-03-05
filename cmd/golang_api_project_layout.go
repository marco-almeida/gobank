package main

import (
	"flag"

	"github.com/marco-almeida/golang-api-project-layout/internal/rest"
	"github.com/marco-almeida/golang-api-project-layout/pkg/logger"
)

func main() {
	listenAddr := flag.String("listen-addr", ":3000", "server listen address")
	flag.Parse()
	logger := logger.New("logs/main.log", true)

	server := rest.NewAPIServer(*listenAddr, logger)
	server.Serve()
}
