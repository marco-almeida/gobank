package storage

import (
	t "github.com/marco-almeida/golang-api-project-layout/internal/types"
)

type Storer interface {
	CreateUser(u *t.RegisterUserRequest) (*t.User, error)
}
