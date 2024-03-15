package rest

import (
	"net/http"

	"github.com/marco-almeida/gobank/internal/storage"
	"github.com/sirupsen/logrus"
)

type APIServer struct {
	addr  string
	log   *logrus.Logger
	store storage.Storer
}

func NewAPIServer(addr string, logger *logrus.Logger, store storage.Storer) *APIServer {
	return &APIServer{
		addr:  addr,
		log:   logger,
		store: store,
	}
}

func (s *APIServer) Serve() {
	router := http.NewServeMux()

	// projectService := NewProjectService(s.store)
	// projectService.RegisterRoutes(subrouter)

	userService := NewUserService(s.log, s.store)
	userService.RegisterRoutes(router)

	// tasksService := NewTasksService(s.store)
	// tasksService.RegisterRoutes(subrouter)

	s.log.Info("Starting the API server at", s.addr)
	loggingMiddleware := LoggingMiddleware(s.log)
	loggedRouter := loggingMiddleware(router)

	s.log.Fatal(http.ListenAndServe(s.addr, loggedRouter))
}
