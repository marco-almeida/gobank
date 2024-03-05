package main

import (
	"github.com/marco-almeida/go-api-structure/internal/rest"
	"github.com/marco-almeida/go-api-structure/pkg/logger"
)

func main() {
	log := logger.MakeLogger("logs/main.log", true)
	log.Info("hello world!")
	log.Fatal("asodasd")

	server := rest.NewAPIServer(":3000")
	server.Serve()
}
