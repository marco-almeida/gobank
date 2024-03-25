package rest

import (
	"net/http"
	"time"

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
	srv := &http.Server{
		Addr:         s.addr,
		Handler:      http.NewServeMux(),
		ReadTimeout:  time.Hour,
		WriteTimeout: time.Hour,
	}

	userService := NewUserService(s.log, s.store)
	userService.RegisterRoutes(srv.Handler.(*http.ServeMux))

	accountsService := NewAccountsService(s.log, s.store)
	accountsService.RegisterRoutes(srv.Handler.(*http.ServeMux))

	s.log.Info("Starting the API server at", s.addr)
	loggingMiddleware := LoggingMiddleware(s.log)
	srv.Handler = loggingMiddleware(RateLimiterMiddleware(srv.Handler))

	s.log.Fatal(srv.ListenAndServe())
}
