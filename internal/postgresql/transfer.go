package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

// TransferRepository represents the repository used for interacting with Transfer records.
type TransferRepository struct {
	q db.Store
}

// NewTransferRepository instantiates the Transfer repository.
func NewTransferRepository(connPool *pgxpool.Pool) *TransferRepository {
	return &TransferRepository{
		q: db.NewStore(connPool),
	}
}

func (transferRepo *TransferRepository) CreateTx(context context.Context, arg db.TransferTxParams) (db.TransferTxResult, error) {
	return transferRepo.q.TransferTx(context, arg)
}
