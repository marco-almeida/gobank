package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marco-almeida/mybank/internal"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

// VerifyEmailRepository represents the repository used for interacting with VerifyEmail records.
type VerifyEmailRepository struct {
	q db.Store
}

// NewVerifyEmailRepository instantiates the VerifyEmail repository.
func NewVerifyEmailRepository(connPool *pgxpool.Pool) *VerifyEmailRepository {
	return &VerifyEmailRepository{
		q: db.NewStore(connPool),
	}
}

func (verifyEmailRepo *VerifyEmailRepository) Create(ctx context.Context, arg db.CreateVerifyEmailParams) (db.VerifyEmail, error) {
	ver, err := verifyEmailRepo.q.CreateVerifyEmail(ctx, arg)
	if err != nil {
		return db.VerifyEmail{}, internal.DBErrorToInternal(err)
	}

	return ver, nil
}

func (verifyEmailRepo *VerifyEmailRepository) Verify(ctx context.Context, arg db.VerifyEmailTxParams) (db.VerifyEmailTxResult, error) {
	res, err := verifyEmailRepo.q.VerifyEmailTx(ctx, arg)
	if err != nil {
		return db.VerifyEmailTxResult{}, internal.DBErrorToInternal(err)
	}

	return res, nil
}
