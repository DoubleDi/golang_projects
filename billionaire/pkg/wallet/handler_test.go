package wallet

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DoubleDi/golang_projects/billionaire/pkg/database"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type BalanceHandlerSuite struct {
	suite.Suite
	repo    *database.MockWalletHistoryRepository
	handler *BalanceHandler
}

func (s *BalanceHandlerSuite) SetupTest() {
	s.repo = &database.MockWalletHistoryRepository{}
	s.handler = NewHandler(s.repo)
}

func (s *BalanceHandlerSuite) TestAddBalance() {
	testCases := []struct {
		name         string
		request      *http.Request
		saveError    error
		expectedCode int
		mock         bool
	}{
		{
			name:         "invalid JSON",
			request:      httptest.NewRequest(http.MethodPost, "/billionaire", bytes.NewBuffer([]byte("{"))),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "negative amount",
			request:      httptest.NewRequest(http.MethodPost, "/billionaire", bytes.NewBuffer([]byte(`{"amount": -1, "datetime": "2022-01-28T14:10:00+00:00"}`))),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "ok",
			request:      httptest.NewRequest(http.MethodPost, "/billionaire", bytes.NewBuffer([]byte(`{"amount": 100, "datetime": "2022-01-28T14:10:00+00:00"}`))),
			expectedCode: http.StatusOK,
			mock:         true,
		},
		{
			name:         "error",
			request:      httptest.NewRequest(http.MethodPost, "/billionaire", bytes.NewBuffer([]byte(`{"amount": 100, "datetime": "2022-01-28T14:10:00+00:00"}`))),
			saveError:    errors.New("some error"),
			expectedCode: http.StatusInternalServerError,
			mock:         true,
		},
	}
	for _, tc := range testCases {
		if tc.mock {
			s.repo.On("AddBalance", context.Background(), mock.Anything).Return(tc.saveError).Once()
		}
		recorder := httptest.NewRecorder()
		s.handler.AddBalance(recorder, tc.request)
		s.EqualValues(tc.expectedCode, recorder.Code, tc.name)
	}
	s.repo.AssertExpectations(s.T())
}

func (s *BalanceHandlerSuite) TestGetBalances() {
	balances := []database.Balance{
		{
			DateTime: time.Date(2019, 10, 5, 12, 0, 0, 0, time.UTC),
			Amount:   decimal.New(1, 0),
		},
	}
	testCases := []struct {
		name         string
		request      *http.Request
		saveError    error
		expectedCode int
		result       string
		mock         bool
	}{
		{
			name:         "from before to",
			request:      httptest.NewRequest(http.MethodGet, "/billionaire?startDatetime=2022-01-28T14:10:00%2B03:00&endDatetime=2022-01-27T14:10:00%2B00:00", nil),
			expectedCode: http.StatusBadRequest,
			result:       "startDatetime after endDatetime\n",
		},
		{
			name:         "invalid from",
			request:      httptest.NewRequest(http.MethodGet, "/billionaire?startDatetime=2022-01-2", nil),
			expectedCode: http.StatusBadRequest,
			result:       "invalid startDatetime value\n",
		},
		{
			name:         "invalid to",
			request:      httptest.NewRequest(http.MethodGet, "/billionaire?endDatetime=2022-01-2", nil),
			expectedCode: http.StatusBadRequest,
			result:       "invalid endDatetime value\n",
		},
		{
			name:         "ok",
			request:      httptest.NewRequest(http.MethodPost, "/billionaire", nil),
			expectedCode: http.StatusOK,
			mock:         true,
			result:       `[{"datetime":"2019-10-05T12:00:00Z","amount":"1"}]` + "\n",
		},
	}
	for _, tc := range testCases {
		if tc.mock {
			s.repo.On("GetBalances", context.Background(), mock.Anything, mock.Anything).Return(balances, tc.saveError).Once()
		}
		recorder := httptest.NewRecorder()
		s.handler.GetBalances(recorder, tc.request)
		s.EqualValues(tc.expectedCode, recorder.Code, tc.name)
		s.EqualValues(tc.result, recorder.Body.String(), tc.name)
	}
	s.repo.AssertExpectations(s.T())
}
func TestBalanceHandlerSuite(t *testing.T) {
	suite.Run(t, new(BalanceHandlerSuite))
}
