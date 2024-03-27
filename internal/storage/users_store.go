package storage

import (
	"database/sql"

	t "github.com/marco-almeida/gobank/internal/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserStore interface {
	Create(u *t.User) error
	GetAll(limit, offset int64) ([]t.User, error)
	DeleteByID(int64) error
	GetByEmail(string) (t.User, error)
	UpdateByID(int64, *t.User) error
	PartialUpdateByID(int64, *t.User) error
	GetByID(int64) (t.User, error)
}

type UsersPostgresStorage struct {
	log *logrus.Logger
	db  *sql.DB
}

func NewUsersPostgresStorage(connStr string, log *logrus.Logger) *UsersPostgresStorage {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Connected to Postgres")

	return &UsersPostgresStorage{log: log, db: db}
}

func (s *UsersPostgresStorage) Init() error {
	err := s.createTable()
	if err != nil {
		return err
	}
	return nil
}

func (s *UsersPostgresStorage) createTable() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY NOT NULL,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)

	return err
}

func (s *UsersPostgresStorage) Create(u *t.User) error {
	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`INSERT INTO users (first_name, last_name, email, password) VALUES ($1, $2, $3, $4)`, u.FirstName, u.LastName, u.Email, string(hashedPassword))

	return err
}

func (s *UsersPostgresStorage) GetAll(limit, offset int64) ([]t.User, error) {
	return nil, nil
}

func (s *UsersPostgresStorage) DeleteByID(int64) error {
	return nil
}

func (s *UsersPostgresStorage) GetByEmail(string) (t.User, error) {
	return t.User{}, nil
}

func (s *UsersPostgresStorage) UpdateByID(int64, *t.User) error {
	return nil
}
func (s *UsersPostgresStorage) PartialUpdateByID(int64, *t.User) error {
	return nil
}
func (s *UsersPostgresStorage) GetByID(int64) (t.User, error) {
	return t.User{}, nil
}
