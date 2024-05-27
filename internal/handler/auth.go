package handler

import (
	"context"

	"github.com/marco-almeida/mybank/internal/postgresql/db"
	"github.com/marco-almeida/mybank/internal/service"
)

// AuthService defines the methods that the auth handler will use
type AuthService interface {
	Create(ctx context.Context, user service.CreateUserParams) (db.User, error)
	Login(ctx context.Context, req service.LoginUserParams) (service.LoginUserResponse, error)
}
