package database

import (
	"regexp"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type PostgresTestSuite struct {
	suite.Suite
	db   *sqlx.DB
	mock sqlmock.Sqlmock
	repo *TransactionPostgresRepository
}

func (s *PostgresTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	s.db = sqlx.NewDb(db, "mock")
	s.mock = mock
	s.repo = NewTransactionPostgresRepository(s.db)
}

func (s *PostgresTestSuite) TearDownTest() {
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

func (s *PostgresTestSuite) TestApplyTransactions() {
	testCases := []struct {
		balance         *Balance
		transactions    []*Transaction
		reverse         bool
		expectedBalance *Balance
	}{
		{
			balance: &Balance{Amount: decimal.New(0, 0)},
			transactions: []*Transaction{
				{State: StateWin, Amount: decimal.New(10, 0)},
				{State: StateLost, Amount: decimal.New(5, 0)},
				{State: "unknown", Amount: decimal.New(5, 0)},
			},
			expectedBalance: &Balance{Amount: decimal.New(5, 0)},
		},
		{
			balance: &Balance{Amount: decimal.New(0, 0)},
			transactions: []*Transaction{
				{State: StateWin, Amount: decimal.New(10, 0)},
				{State: StateLost, Amount: decimal.New(5, 0)},
				{State: "unknown", Amount: decimal.New(5, 0)},
			},
			expectedBalance: &Balance{Amount: decimal.New(-5, 0)},
			reverse:         true,
		},
	}

	for i, tc := range testCases {
		s.EqualValues(tc.expectedBalance, applyTransactions(tc.balance, tc.reverse, tc.transactions...), "%d", i)
	}
}

func (s *PostgresTestSuite) TestSaveTransaction() {
	t := time.Now()
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM balance limit 1")).WillReturnRows(
		sqlmock.NewRows([]string{"amount", "updated_at"}).AddRow(decimal.New(10, 0), t),
	)
	s.mock.ExpectExec(regexp.QuoteMeta("INSERT INTO transactions (id, state, amount) VALUES (?, ?, ?)")).WithArgs(
		"123", StateWin, decimal.New(10, 0),
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)
	s.mock.ExpectExec(regexp.QuoteMeta("UPDATE balance SET amount = ?, updated_at = ?")).WithArgs(
		decimal.New(20, 0), t,
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)
	s.mock.ExpectCommit()
	s.Require().NoError(s.repo.SaveTransaction(&Transaction{ID: "123", State: StateWin, Amount: decimal.New(10, 0)}))
}

func (s *PostgresTestSuite) TestRemoveOddTransactions() {
	t := time.Now()
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM balance limit 1")).WillReturnRows(
		sqlmock.NewRows([]string{"amount", "updated_at"}).AddRow(decimal.New(21, 0), t),
	)
	rows := sqlmock.NewRows([]string{"id", "amount", "created_at", "state"})
	for i := 0; i < 20; i++ {
		rows.AddRow("xxx", decimal.New(1, 0), t, StateWin)
	}
	s.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM transactions WHERE created_at > $1 AND reverted = $2 ORDER BY created_at DESC")).WillReturnRows(
		rows,
	)
	s.mock.ExpectExec(regexp.QuoteMeta("UPDATE transactions SET reverted = $1 WHERE id IN ($2,$3,$4,$5,$6,$7,$8,$9,$10,$11)")).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)
	s.mock.ExpectExec(regexp.QuoteMeta("UPDATE balance SET amount = ?, updated_at = ?")).WithArgs(
		decimal.New(11, 0), t,
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)
	s.mock.ExpectCommit()
	s.Require().NoError(s.repo.RemoveOddTransactions())
}

func (s *PostgresTestSuite) TestRemoveOddTransactionsLessThan20() {
	t := time.Now()
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM balance limit 1")).WillReturnRows(
		sqlmock.NewRows([]string{"amount", "updated_at"}).AddRow(decimal.New(21, 0), t),
	)
	s.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM transactions WHERE created_at > $1 AND reverted = $2 ORDER BY created_at DESC")).WillReturnRows(
		sqlmock.NewRows([]string{"id", "amount", "created_at", "state"}).AddRow("xxx", decimal.New(1, 0), t, StateLost),
	)
	s.mock.ExpectRollback()
	s.Require().NoError(s.repo.RemoveOddTransactions())
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}
