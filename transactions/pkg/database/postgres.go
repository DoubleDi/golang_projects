package database

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// State types
const (
	StateWin  = "win"
	StateLost = "lost"
)

// Errors on handling transactions and balance
var (
	ErrNegativeBalance = errors.New("balance can't be negative")
)

// Transaction represents the transaction table in db
type Transaction struct {
	ID        string          `db:"id"`
	State     string          `db:"state"`
	Amount    decimal.Decimal `db:"amount"`
	CreatedAt time.Time       `db:"created_at"`
	Reverted  bool            `db:"reverted"`
}

// Balance trepresents the balance table in db
type Balance struct {
	Amount    decimal.Decimal `db:"amount"`
	UpdatedAt time.Time       `db:"updated_at"`
}

// TransactionPostgresRepository represents the postgres TransactionsRepository
// implementation
type TransactionPostgresRepository struct {
	db *sqlx.DB
}

// NewTransactionPostgresRepository returns a new repository instance
func NewTransactionPostgresRepository(db *sqlx.DB) *TransactionPostgresRepository {
	return &TransactionPostgresRepository{
		db: db,
	}
}

func applyTransactions(balance *Balance, reverse bool, transactions ...*Transaction) *Balance {
	for _, transaction := range transactions {
		log.Printf("%#v", transaction)
		if reverse && transaction.State == StateWin || !reverse && transaction.State == StateLost {
			balance.Amount = balance.Amount.Sub(transaction.Amount)
			continue
		}
		if reverse && transaction.State == StateLost || !reverse && transaction.State == StateWin {
			balance.Amount = balance.Amount.Add(transaction.Amount)
			continue
		}
	}
	return balance
}

// SaveTransaction saves a transaction to db
func (r *TransactionPostgresRepository) SaveTransaction(t *Transaction) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	balance, err := r.getBalance(tx)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "error on getting balance")
	}
	if t.State == StateLost && t.Amount.GreaterThan(balance.Amount) {
		return ErrNegativeBalance
	}

	_, err = tx.NamedExec(
		"INSERT INTO transactions (id, state, amount) VALUES (:id, :state, :amount)",
		map[string]interface{}{"id": t.ID, "state": t.State, "amount": t.Amount},
	)
	if err != nil {
		return errors.Wrap(err, "error on inserting transaction")
	}

	balance = applyTransactions(balance, false, t)
	if err := r.saveBalance(tx, balance); err != nil {
		return errors.Wrap(err, "error on saving balance")
	}
	return tx.Commit()
}

func (r *TransactionPostgresRepository) getBalance(db sqlx.Ext) (*Balance, error) {
	var balance Balance
	err := sqlx.Get(db, &balance, "SELECT * FROM balance limit 1")
	return &balance, err
}

func (r *TransactionPostgresRepository) saveBalance(db sqlx.Ext, b *Balance) error {
	_, err := sqlx.NamedExec(
		db,
		"UPDATE balance SET amount = :amount, updated_at = :updated_at",
		map[string]interface{}{"amount": b.Amount, "updated_at": b.UpdatedAt},
	)
	return err
}

func (r *TransactionPostgresRepository) getLastTransactions(db sqlx.Ext, after time.Time) ([]*Transaction, error) {
	var transactions []*Transaction
	err := sqlx.Select(db, &transactions, "SELECT * FROM transactions WHERE created_at > $1 AND reverted = $2 ORDER BY created_at DESC", after, false)
	return transactions, err
}

func (r *TransactionPostgresRepository) removeTransactions(db sqlx.Ext, transactions []*Transaction) error {
	args := make([]interface{}, 0, len(transactions)+1)
	args = append(args, true)
	holders := make([]string, 0, len(transactions))
	for i, transaction := range transactions {
		args = append(args, transaction.ID)
		holders = append(holders, "$"+strconv.Itoa(i+2))
	}
	log.Printf("Deleting transactions %v", args)
	// using soft deleting here so we keep the ids in database, so we can check the uniqueness of the
	// incoming transactions
	_, err := db.Exec("UPDATE transactions SET reverted = $1 WHERE id IN ("+strings.Join(holders, ",")+")", args...)
	return err
}

// RemoveOddTransactions reverts last 10 odd transactions from db
func (r *TransactionPostgresRepository) RemoveOddTransactions() error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	balance, err := r.getBalance(tx)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "error on getting balance")
	}
	transactions, err := r.getLastTransactions(tx, balance.UpdatedAt)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "error on getting last transactions")
	}
	// 10 latest odd records should be processed, so we need to check
	// for at least 20 rows
	if len(transactions) < 20 {
		tx.Rollback()
		log.Println("Less than 10 odd rows found. Not deleting anything")
		return nil
	}

	toDelete := make([]*Transaction, 0, 10)
	for i, transaction := range transactions {
		if (i+1)%2 == 0 {
			toDelete = append(toDelete, transaction)
		}
	}

	if err := r.removeTransactions(tx, toDelete); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "error on removing transactions")
	}

	balance = applyTransactions(balance, true, toDelete...)
	// setting update time to the last latest transactions
	// will be checking 20 more transactions from that time next time
	balance.UpdatedAt = transactions[0].CreatedAt

	if err := r.saveBalance(tx, balance); err != nil {
		return errors.Wrap(err, "error on saving balance")
	}
	return tx.Commit()
}
