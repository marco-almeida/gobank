package rest

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

type APIServer struct {
	addr string
	log  *logrus.Logger
}

func NewAPIServer(addr string, logger *logrus.Logger) *APIServer {
	return &APIServer{
		addr: addr,
		log:  logger,
	}
}

func (s *APIServer) Serve() {
	router := http.NewServeMux()
	// router := mux.NewRouter()
	// subrouter := router.PathPrefix("/api/v1").Subrouter()

	// projectService := NewProjectService(s.store)
	// projectService.RegisterRoutes(subrouter)

	userService := NewUserService(s.log)
	userService.RegisterRoutes(router)

	// tasksService := NewTasksService(s.store)
	// tasksService.RegisterRoutes(subrouter)

	s.log.Info("Starting the API server at", s.addr)
	loggingMiddleware := LoggingMiddleware(s.log)
	loggedRouter := loggingMiddleware(router)

	s.log.Fatal(http.ListenAndServe(s.addr, loggedRouter))
}
