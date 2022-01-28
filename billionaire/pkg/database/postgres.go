package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type Balance struct {
	DateTime time.Time       `db:"datetime" json:"datetime"`
	Amount   decimal.Decimal `db:"amount"   json:"amount"`
}

type WalletHistoryPostgresRepository struct {
	db *sqlx.DB
}

func NewWalletHistoryPostgresRepository(db *sqlx.DB) *WalletHistoryPostgresRepository {
	return &WalletHistoryPostgresRepository{
		db: db,
	}
}

func (r *WalletHistoryPostgresRepository) AddBalance(ctx context.Context, b *Balance) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	hour, err := r.getLatestTime(ctx, tx)
	if err != nil {
		return err
	}
	if hour.After(b.DateTime) {
		if err := r.updateFutureAmounts(ctx, tx, b); err != nil {
			return err
		}

	}
	if err := r.upsertAmount(ctx, tx, b); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *WalletHistoryPostgresRepository) getLatestTime(ctx context.Context, tx *sqlx.Tx) (time.Time, error) {
	var t time.Time
	err := tx.GetContext(ctx, &t, "SELECT datetime FROM wallet_hour_history ORDER BY datetime DESC LIMIT 1")
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return time.Time{}, nil
	case err == nil:
		return t, nil
	default:
		return time.Time{}, err
	}
}

func (r *WalletHistoryPostgresRepository) updateFutureAmounts(ctx context.Context, tx *sqlx.Tx, b *Balance) error {
	_, err := tx.NamedExecContext(
		ctx,
		"UPDATE wallet_hour_history SET amount = amount + :amount WHERE datetime > :datetime",
		map[string]interface{}{"amount": b.Amount, "datetime": b.DateTime},
	)
	return err
}

func (r *WalletHistoryPostgresRepository) upsertAmount(ctx context.Context, tx *sqlx.Tx, b *Balance) error {
	_, err := tx.NamedExecContext(
		ctx,
		"INSERT INTO wallet_hour_history AS w (datetime, amount) VALUES (:datetime, :amount) ON CONFLICT (datetime) DO UPDATE SET amount = w.amount + EXCLUDED.amount",
		map[string]interface{}{"amount": b.Amount, "datetime": b.DateTime},
	)
	return err
}

func (r *WalletHistoryPostgresRepository) GetBalances(ctx context.Context, from, to time.Time) ([]Balance, error) {
	currentBalances := make([]Balance, 0, 100)
	query, args, err := r.db.BindNamed(
		"SELECT datetime, amount FROM wallet_hour_history WHERE datetime >= :from AND datetime <= :to ORDER BY datetime",
		map[string]interface{}{"from": from, "to": to},
	)
	if err != nil {
		return nil, err
	}

	err = r.db.SelectContext(
		ctx,
		&currentBalances,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}

	balances := make([]Balance, 0, len(currentBalances)+1)
	query, args, err = r.db.BindNamed(
		"SELECT datetime, amount FROM wallet_hour_history WHERE datetime < :from ORDER BY datetime DESC LIMIT 1",
		map[string]interface{}{"from": from},
	)
	if err != nil {
		return nil, err
	}

	err = r.db.SelectContext(
		ctx,
		&balances,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}
	balances = append(balances, currentBalances...)

	if len(balances) == 0 {
		balances = append(balances, Balance{})
	}

	return r.fillBalances(balances, from, to), nil
}

// 12
// 0

func (r *WalletHistoryPostgresRepository) fillBalances(balances []Balance, from, to time.Time) []Balance {
	if len(balances) == 0 {
		return balances
	}
	fmt.Println(balances)

	prevBalance := &balances[0]
	hourBalance := make(map[string]*Balance, len(balances))
	for i := range balances {
		hourBalance[balances[i].DateTime.UTC().String()] = &balances[i]
	}
	result := make([]Balance, 0, int(to.Sub(from).Hours()+1))
	for t := from.UTC(); t.Before(to) || t.Equal(to); t = t.Add(time.Hour) {
		if balance, ok := hourBalance[t.UTC().String()]; ok {
			prevBalance = balance
			b := *balance
			b.DateTime = t
			result = append(result, b)
			continue
		}
		b := *prevBalance
		b.DateTime = t
		result = append(result, b)
	}

	return result
}
