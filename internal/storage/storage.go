package storage

import (
	t "github.com/marco-almeida/golang-api-project-layout/internal/types"
)

type Storer interface {
	// returns given id to user
	CreateUser(u *t.RegisterUserRequest) error
	GetAllUsers() ([]t.User, error)
	DeleteUserByID(int64) error
	GetUserByEmail(string) (t.User, error)
}
