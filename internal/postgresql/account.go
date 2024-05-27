package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

// AccountRepository represents the repository used for interacting with Account records.
type AccountRepository struct {
	q db.Store
}

// NewAccountRepository instantiates the Account repository.
func NewAccountRepository(connPool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{
		q: db.NewStore(connPool),
	}
}

func (accountRepo *AccountRepository) Create(ctx context.Context, arg db.CreateAccountParams) (db.Account, error) {
	return accountRepo.q.CreateAccount(ctx, arg)
}

func (accountRepo *AccountRepository) Get(ctx context.Context, id int64) (db.Account, error) {
	return accountRepo.q.GetAccount(ctx, id)
}

func (accountRepo *AccountRepository) List(ctx context.Context, arg db.ListAccountsParams) ([]db.Account, error) {
	return accountRepo.q.ListAccounts(ctx, arg)
}
