package app

import (
	"github.com/DoubleDi/golang_projects/transactions/pkg/transaction"
	"github.com/go-playground/validator"
)

// NewValidator returns a new argument validator instance
func NewValidator() *validator.Validate {
	validate := validator.New()
	validate.RegisterValidation("validate_state", transaction.ValidateState)
	return validate
}
