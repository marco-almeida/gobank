package service

import (
	"context"

	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

// TransferRepository defines the methods that any Transfer repository should implement.
type TransferRepository interface {
	CreateTx(context context.Context, arg db.TransferTxParams) (db.TransferTxResult, error)
}

// TransferService defines the application service in charge of interacting with Transfers.
type TransferService struct {
	repo TransferRepository
}

// NewTransferService creates a new User service.
func NewTransferService(repo TransferRepository) *TransferService {
	return &TransferService{
		repo: repo,
	}
}

func (s *TransferService) CreateTx(context context.Context, arg db.TransferTxParams) (db.TransferTxResult, error) {
	return s.repo.CreateTx(context, arg)
}
