package postgresql

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marco-almeida/mybank/internal"
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
	session, err := sessionRepo.q.CreateSession(ctx, arg)
	if err != nil {
		return db.Session{}, internal.DBErrorToInternal(err)
	}

	return session, nil
}

func (sessionRepo *SessionRepository) Get(ctx context.Context, id uuid.UUID) (db.Session, error) {
	session, err := sessionRepo.q.GetSession(ctx, id)
	if err != nil {
		return db.Session{}, internal.DBErrorToInternal(err)
	}

	return session, nil
}
