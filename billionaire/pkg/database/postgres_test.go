package database

import (
	"context"
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
	repo *WalletHistoryPostgresRepository
}

func (s *PostgresTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	s.db = sqlx.NewDb(db, "mock")
	s.mock = mock
	s.repo = NewWalletHistoryPostgresRepository(s.db)
}

func (s *PostgresTestSuite) TearDownTest() {
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

func (s *PostgresTestSuite) TestAddBalanceNewest() {
	t := time.Date(2019, 10, 5, 13, 0, 0, 0, time.UTC)
	t2 := t.AddDate(0, 0, -1)
	balance := &Balance{
		DateTime: time.Date(2019, 10, 5, 13, 0, 0, 0, time.UTC),
		Amount:   decimal.New(2000, 1),
	}
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta("SELECT datetime FROM wallet_hour_history ORDER BY datetime DESC LIMIT 1")).WillReturnRows(
		sqlmock.NewRows([]string{"datetime"}).AddRow(t2),
	)
	s.mock.ExpectExec(regexp.QuoteMeta("INSERT INTO wallet_hour_history AS w (datetime, amount) VALUES (?, ?) ON CONFLICT (datetime) DO UPDATE SET amount = w.amount + EXCLUDED.amount")).WithArgs(
		balance.DateTime, balance.Amount,
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)
	s.mock.ExpectCommit()
	s.Require().NoError(s.repo.AddBalance(context.Background(), balance))
}

func (s *PostgresTestSuite) TestAddBalanceOld() {
	t := time.Date(2019, 10, 5, 13, 0, 0, 0, time.UTC)
	t2 := t.AddDate(0, 0, 1)
	balance := &Balance{
		DateTime: time.Date(2019, 10, 5, 13, 0, 0, 0, time.UTC),
		Amount:   decimal.New(2000, 1),
	}
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta("SELECT datetime FROM wallet_hour_history ORDER BY datetime DESC LIMIT 1")).WillReturnRows(
		sqlmock.NewRows([]string{"datetime"}).AddRow(t2),
	)
	s.mock.ExpectExec(regexp.QuoteMeta("UPDATE wallet_hour_history SET amount = amount + ? WHERE datetime > ?")).WithArgs(
		balance.Amount, balance.DateTime,
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)
	s.mock.ExpectExec(regexp.QuoteMeta("INSERT INTO wallet_hour_history AS w (datetime, amount) VALUES (?, ?) ON CONFLICT (datetime) DO UPDATE SET amount = w.amount + EXCLUDED.amount")).WithArgs(
		balance.DateTime, balance.Amount,
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)
	s.mock.ExpectCommit()
	s.Require().NoError(s.repo.AddBalance(context.Background(), balance))
}

func (s *PostgresTestSuite) TestGetBalances() {
	from := time.Date(2019, 10, 5, 12, 0, 0, 0, time.UTC)
	to := time.Date(2019, 10, 5, 16, 0, 0, 0, time.UTC)
	expected := []Balance{
		{
			DateTime: time.Date(2019, 10, 5, 12, 0, 0, 0, time.UTC),
			Amount:   decimal.New(1, 0),
		},
		{
			DateTime: time.Date(2019, 10, 5, 13, 0, 0, 0, time.UTC),
			Amount:   decimal.New(1, 0),
		},
		{
			DateTime: time.Date(2019, 10, 5, 14, 0, 0, 0, time.UTC),
			Amount:   decimal.New(101, 0),
		},
		{
			DateTime: time.Date(2019, 10, 5, 15, 0, 0, 0, time.UTC),
			Amount:   decimal.New(201, 0),
		},
		{
			DateTime: time.Date(2019, 10, 5, 16, 0, 0, 0, time.UTC),
			Amount:   decimal.New(201, 0),
		},
	}

	s.mock.ExpectQuery(
		regexp.QuoteMeta("SELECT datetime, amount FROM wallet_hour_history WHERE datetime >= ? AND datetime <= ? ORDER BY datetime"),
	).WithArgs(from, to).WillReturnRows(
		sqlmock.NewRows([]string{"datetime", "amount"}).AddRow(
			time.Date(2019, 10, 5, 14, 0, 0, 0, time.UTC), decimal.New(101, 0),
		).AddRow(
			time.Date(2019, 10, 5, 15, 0, 0, 0, time.UTC), decimal.New(201, 0),
		),
	)
	s.mock.ExpectQuery(
		regexp.QuoteMeta("SELECT datetime, amount FROM wallet_hour_history WHERE datetime < ? ORDER BY datetime DESC LIMIT 1"),
	).WithArgs(from).WillReturnRows(
		sqlmock.NewRows([]string{"datetime", "amount"}).AddRow(
			time.Date(2019, 10, 5, 11, 0, 0, 0, time.UTC), decimal.New(1, 0),
		),
	)
	balances, err := s.repo.GetBalances(context.Background(), from, to)
	s.Require().NoError(err)
	s.Require().EqualValues(len(expected), len(balances))
	s.EqualValues(expected, balances)
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}
