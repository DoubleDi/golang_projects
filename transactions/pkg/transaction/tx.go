package transaction

import (
	"github.com/DoubleDi/golang_projects/transactions/pkg/database"
	"github.com/go-playground/validator"
	"github.com/shopspring/decimal"
)

var availableStates = map[string]struct{}{
	database.StateWin:  {},
	database.StateLost: {},
}

// ValidateState - validates states
func ValidateState(fl validator.FieldLevel) bool {
	_, ok := availableStates[fl.Field().String()]
	return ok
}

// Possible Source-Type header values
const (
	SourceTypeGame    = "game"
	SourceTypeServer  = "server"
	SourceTypePayment = "payment"
)

var availableSourceTypes = map[string]struct{}{
	SourceTypeGame:    {},
	SourceTypeServer:  {},
	SourceTypePayment: {},
}

func validateSourceType(sourceType string) bool {
	_, ok := availableSourceTypes[sourceType]
	return ok
}

// HandleTransactionArgs - arguments for HandleTransaction method
type HandleTransactionArgs struct {
	State         string          `json:"state" validate:"required,validate_state"`
	Amount        decimal.Decimal `json:"amount" validate:"required"`
	TransactionID string          `json:"transactionId" validate:"required"`
}
