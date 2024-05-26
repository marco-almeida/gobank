package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

// UserRepository represents the repository used for interacting with User records.
type UserRepository struct {
	q db.Store
}

// NewUser instantiates the User repository.
func NewUserRepository(connPool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		q: db.NewStore(connPool),
	}
}

func (userRepo *UserRepository) Create(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return userRepo.q.CreateUser(ctx, arg)
}
