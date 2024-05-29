package internal

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUniqueConstraintViolation     = errors.New("unique constraint violation")
	ErrForeignKeyConstraintViolation = errors.New("foreign key constraint violation")
	ErrNoRows                        = errors.New("no rows in result set")
	ErrUnauthorized                  = errors.New("unauthorized")
	ErrInvalidToken                  = errors.New("invalid token")
	ErrInvalidCredentials            = errors.New("wrong password")
	ErrInvalidParams                 = errors.New("invalid params")
	ErrForbidden                     = errors.New("forbidden")
	ErrInternal                      = errors.New("internal error")
)

// db error to internal error
func DBErrorToInternal(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %s", ErrNoRows, err.Error())
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return fmt.Errorf("%w: %s", ErrForeignKeyConstraintViolation, pgErr.Detail)
		case "23505":
			return fmt.Errorf("%w: %s", ErrUniqueConstraintViolation, pgErr.Detail)
		default:
			return fmt.Errorf("%w: %s", ErrInternal, pgErr.Detail)
		}
	}
	return err
}
