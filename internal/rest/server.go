package rest

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type APIServer struct {
	addr string
}

func NewAPIServer(addr string) *APIServer {
	return &APIServer{
		addr: addr,
	}
}

func (s *APIServer) Serve() {
	router := http.NewServeMux()
	// router := mux.NewRouter()
	// subrouter := router.PathPrefix("/api/v1").Subrouter()

	// projectService := NewProjectService(s.store)
	// projectService.RegisterRoutes(subrouter)

	// userService := NewUserService(s.store)
	// userService.RegisterRoutes(subrouter)

	// tasksService := NewTasksService(s.store)
	// tasksService.RegisterRoutes(subrouter)

	// TODO: use same logger for all services

	log.Println("Starting the API server at", s.addr)

	log.Fatal(http.ListenAndServe(s.addr, router))
}
