package transaction

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoubleDi/golang_projects/transactions/pkg/database"
	"github.com/go-playground/validator"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TransactionHandlerSuite struct {
	suite.Suite
	repo    *database.MockTransactionRepository
	handler *TransactionHandler
}

func (s *TransactionHandlerSuite) SetupTest() {
	s.repo = &database.MockTransactionRepository{}
	validate := validator.New()
	validate.RegisterValidation("validate_state", ValidateState)
	s.handler = NewHandler(s.repo, validate)
}

func (s *TransactionHandlerSuite) TestHandleTransaction() {
	testCases := []struct {
		name         string
		request      *http.Request
		sourceType   string
		saveError    error
		expectedCode int
		mock         bool
	}{
		{
			name:         "invalid Source-Type",
			request:      httptest.NewRequest(http.MethodPost, "/transactions", nil),
			sourceType:   "invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid JSON",
			request:      httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer([]byte("{"))),
			sourceType:   "game",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid State",
			request:      httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer([]byte(`{"state": "xxx"}`))),
			sourceType:   "game",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid State",
			request:      httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer([]byte(`{"state": "xxx"}`))),
			sourceType:   "game",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "ok",
			request:      httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer([]byte(`{"state": "win", "transactionId": "xxx", "amount": "10.1"}`))),
			sourceType:   "game",
			expectedCode: http.StatusOK,
			mock:         true,
		},
		{
			name:         "unique error",
			request:      httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer([]byte(`{"state": "win", "transactionId": "xxx", "amount": "10.1"}`))),
			sourceType:   "game",
			saveError:    pgx.PgError{Code: pgerrcode.UniqueViolation},
			expectedCode: http.StatusConflict,
			mock:         true,
		},
		{
			name:         "error",
			request:      httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer([]byte(`{"state": "win", "transactionId": "xxx", "amount": "10.1"}`))),
			sourceType:   "game",
			saveError:    errors.New("some error"),
			expectedCode: http.StatusBadRequest,
			mock:         true,
		},
	}
	// TODO: amount < 0
	for _, tc := range testCases {
		tc.request.Header.Set("Source-Type", tc.sourceType)
		if tc.mock {
			s.repo.On("SaveTransaction", mock.Anything).Return(tc.saveError).Once()
		}
		recorder := httptest.NewRecorder()
		s.handler.HandleTransaction(recorder, tc.request)
		s.EqualValues(tc.expectedCode, recorder.Code, tc.name)
	}
	s.repo.AssertExpectations(s.T())
}
func TestTransactionhandlerSuite(t *testing.T) {
	suite.Run(t, new(TransactionHandlerSuite))
}
