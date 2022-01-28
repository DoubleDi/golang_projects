package database

import (
	"context"
	"time"
)

type WalletHistoryRepository interface {
	AddBalance(ctx context.Context, b *Balance) error
	GetBalances(ctx context.Context, from, to time.Time) ([]Balance, error)
}
