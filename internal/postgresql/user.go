package postgres

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/marco-almeida/gobank/internal"
)

// User represents the repository used for interacting with User records.
type User struct {
	db *sql.DB
}

// NewUser instantiates the User repository.
func NewUser(db *sql.DB) *User {
	return &User{
		db: db,
	}
}

func (s *User) Init() error {
	return s.createTable()
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

func (s *User) Create(u *internal.User) error {
	_, err := s.db.Exec(`INSERT INTO users (first_name, last_name, email, password) VALUES ($1, $2, $3, $4)`, u.FirstName, u.LastName, u.Email, u.Password)

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

func (s *User) GetAll(limit, offset int64) ([]internal.User, error) {
	return nil, nil
}

func (s *User) DeleteByID(int64) error {
	return nil
}

func (s *User) GetByEmail(string) (internal.User, error) {
	return internal.User{}, nil
}

func (s *User) UpdateByID(int64, *internal.User) error {
	return nil
}
func (s *User) PartialUpdateByID(int64, *internal.User) error {
	return nil
}
func (s *User) GetByID(int64) (internal.User, error) {
	return internal.User{}, nil
}
