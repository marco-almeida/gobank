package storage

import (
	"database/sql"

	_ "github.com/lib/pq"
	t "github.com/marco-almeida/golang-api-project-layout/internal/types"
	"github.com/sirupsen/logrus"
)

type PostgresStorage struct {
	log *logrus.Logger
}

func NewPostgresStorage(connStr string, log *logrus.Logger) *PostgresStorage {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Connected to Postgres")

	return &PostgresStorage{log: log}
}

func (s *PostgresStorage) Init() error {
	s.log.Info("Initializing users table")
	return nil
}

func (s *PostgresStorage) CreateUser(u *t.RegisterUserRequest) (*t.User, error) {
	return &t.User{ID: 3}, nil
}
