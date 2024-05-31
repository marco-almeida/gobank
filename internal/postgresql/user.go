package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marco-almeida/mybank/internal"
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
	user, err := userRepo.q.CreateUser(ctx, arg)
	if err != nil {
		return db.User{}, internal.DBErrorToInternal(err)
	}
	return user, nil
}

func (userRepo *UserRepository) Get(ctx context.Context, username string) (db.User, error) {
	user, err := userRepo.q.GetUser(ctx, username)
	if err != nil {
		return db.User{}, internal.DBErrorToInternal(err)
	}
	return user, nil
}

func (userRepo *UserRepository) CreateWithTx(ctx context.Context, arg db.CreateUserTxParams) (db.CreateUserTxResult, error) {
	res, err := userRepo.q.CreateUserTx(ctx, arg)
	if err != nil {
		return db.CreateUserTxResult{}, internal.DBErrorToInternal(err)
	}
	return res, nil
}
