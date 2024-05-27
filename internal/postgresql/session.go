package postgresql

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

// SessionRepository represents the repository used for interacting with Session records.
type SessionRepository struct {
	q db.Store
}

// NewSessionRepository instantiates the Session repository.
func NewSessionRepository(connPool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{
		q: db.NewStore(connPool),
	}
}

func (sessionRepo *SessionRepository) Create(ctx context.Context, arg db.CreateSessionParams) (db.Session, error) {
	return sessionRepo.q.CreateSession(ctx, arg)
}

func (sessionRepo *SessionRepository) Get(ctx context.Context, id uuid.UUID) (db.Session, error) {
	return sessionRepo.q.GetSession(ctx, id)
}
