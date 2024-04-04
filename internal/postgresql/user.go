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
	rows, err := s.db.Query(`SELECT id, first_name, last_name, email, password, created_at FROM users LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "get all users")
	}

	defer rows.Close()

	var users []internal.User

	for rows.Next() {
		var u internal.User
		err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.CreatedAt)
		if err != nil {
			return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "get all users")
		}

		users = append(users, u)
	}

	return users, nil
}

func (s *User) DeleteByID(id int64) error {
	_, err := s.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "delete user")
	}

	return nil
}

func (s *User) GetByEmail(email string) (internal.User, error) {
	var u internal.User
	err := s.db.QueryRow(`SELECT id, first_name, last_name, email, password, created_at FROM users WHERE email = $1`, email).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.CreatedAt)
	if err != nil {
		// if err of type no rows, return 404
		if errors.Is(err, sql.ErrNoRows) {
			return internal.User{}, internal.WrapErrorf(err, internal.ErrorCodeNotFound, "user not found")
		}
		return internal.User{}, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "get user by email")
	}

	return u, nil
}

func (s *User) UpdateByID(id int64, u *internal.User) (internal.User, error) {
	// password already hashed
	var updatedUser internal.User
	err := s.db.QueryRow(`UPDATE users SET first_name = $1, last_name = $2, email = $3, password = $4 WHERE id = $5 RETURNING id, first_name, last_name, email, password`, u.FirstName, u.LastName, u.Email, u.Password, id).Scan(&updatedUser.ID, &updatedUser.FirstName, &updatedUser.LastName, &updatedUser.Email, &updatedUser.Password)
	if err != nil {
		// if err of type no rows, return 404
		if errors.Is(err, sql.ErrNoRows) {
			return internal.User{}, internal.WrapErrorf(err, internal.ErrorCodeNotFound, "user not found")
		}
		// check if error of type duplicate key
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return internal.User{}, internal.WrapErrorf(err, internal.ErrorCodeDuplicate, "email already in use")
			}
		}
		return internal.User{}, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "update user")
	}

	return updatedUser, nil
}
func (s *User) PartialUpdateByID(id int64, u *internal.User) (internal.User, error) {
	// if password exists, already hashed
	var firstName *string
	var lastName *string
	var email *string
	var password *string

	if u.FirstName != "" {
		firstName = &u.FirstName
	}
	if u.LastName != "" {
		lastName = &u.LastName
	}
	if u.Email != "" {
		email = &u.Email
	}
	if u.Password != "" {
		password = &u.Password
	}

	var updatedUser internal.User
	err := s.db.QueryRow(`UPDATE users SET first_name = COALESCE($1, first_name), last_name = COALESCE($2, last_name), email = COALESCE($3, email), password = COALESCE($4, password) WHERE id = $5 RETURNING id, first_name, last_name, email, password`, firstName, lastName, email, password, id).Scan(&updatedUser.ID, &updatedUser.FirstName, &updatedUser.LastName, &updatedUser.Email, &updatedUser.Password)
	if err != nil {
		// if err of type no rows, return 404
		if errors.Is(err, sql.ErrNoRows) {
			return internal.User{}, internal.WrapErrorf(err, internal.ErrorCodeNotFound, "user not found")
		}
		// check if error of type duplicate key
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return internal.User{}, internal.WrapErrorf(err, internal.ErrorCodeDuplicate, "email already in use")
			}
		}
		return internal.User{}, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "update user")
	}

	return updatedUser, nil
}

func (s *User) GetByID(id int64) (internal.User, error) {
	var u internal.User
	err := s.db.QueryRow(`SELECT id, first_name, last_name, email, password, created_at FROM users WHERE id = $1`, id).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.CreatedAt)
	if err != nil {
		// if err of type no rows, return 404
		if errors.Is(err, sql.ErrNoRows) {
			return internal.User{}, internal.WrapErrorf(err, internal.ErrorCodeNotFound, "user not found")
		}
		return internal.User{}, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "get user by id")
	}

	return u, nil
}
