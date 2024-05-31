package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
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
	return verifyEmailRepo.q.CreateVerifyEmail(ctx, arg)
}
