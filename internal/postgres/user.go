package postgres

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/marco-almeida/gobank/internal"
	"github.com/marco-almeida/gobank/internal/model"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// User represents the repository used for interacting with User records.
type User struct {
	log *logrus.Logger
	db  *sql.DB
}

// NewUser instantiates the User repository.
func NewUser(connStr string, log *logrus.Logger) *User {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Connected to Postgres")

	return &User{log: log, db: db}
}

func (s *User) Init() error {
	err := s.createTable()
	if err != nil {
		return err
	}
	return nil
}

func (s *User) createTable() error {
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

func (s *User) Create(u *model.User) error {
	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`INSERT INTO users (first_name, last_name, email, password) VALUES ($1, $2, $3, $4)`, u.FirstName, u.LastName, u.Email, string(hashedPassword))

	if err != nil {
		var pgErr *pq.Error
		// check if error of type duplicate key
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return internal.WrapErrorf(err, internal.ErrorCodeDuplicate, "email already in use")
			}
		}

		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "create user")
	}

	return err
}

func (s *User) GetAll(limit, offset int64) ([]model.User, error) {
	return nil, nil
}

func (s *User) DeleteByID(int64) error {
	return nil
}

func (s *User) GetByEmail(string) (model.User, error) {
	return model.User{}, nil
}

func (s *User) UpdateByID(int64, *model.User) error {
	return nil
}
func (s *User) PartialUpdateByID(int64, *model.User) error {
	return nil
}
func (s *User) GetByID(int64) (model.User, error) {
	return model.User{}, nil
}
