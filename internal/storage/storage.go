package storage

import (
	"context"
	"ewallet/internal/model"
)

type Storage interface {
	SaveWallet(ctx context.Context, w *model.Wallet) error
	SaveTransaction(ctx context.Context, t *model.Transaction) error
	GetLastTransactions(ctx context.Context, n int) (*[]model.Transaction, error)
	GetBalance(ctx context.Context, w *model.Wallet) error
	RemoveWallet(ctx context.Context, address string)
	RemoveTransaction(ctx context.Context, t *model.Transaction)
	IsExistsWallet(ctx context.Context, w *model.Wallet)
}
