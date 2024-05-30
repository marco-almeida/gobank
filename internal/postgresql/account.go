package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marco-almeida/mybank/internal"
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
	account, err := accountRepo.q.CreateAccount(ctx, arg)
	if err != nil {
		return db.Account{}, internal.DBErrorToInternal(err)
	}
	return account, nil
}

func (accountRepo *AccountRepository) Get(ctx context.Context, id int64) (db.Account, error) {
	account, err := accountRepo.q.GetAccount(ctx, id)
	if err != nil {
		return db.Account{}, internal.DBErrorToInternal(err)
	}
	return account, nil
}

func (accountRepo *AccountRepository) List(ctx context.Context, arg db.ListAccountsParams) ([]db.Account, error) {
	accounts, err := accountRepo.q.ListAccounts(ctx, arg)
	if err != nil {
		return []db.Account{}, internal.DBErrorToInternal(err)
	}
	return accounts, nil
}
