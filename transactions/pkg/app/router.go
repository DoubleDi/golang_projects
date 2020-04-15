package app

import (
	"net/http"
	"os"

	"github.com/DoubleDi/golang_projects/transactions/pkg/database"
	"github.com/DoubleDi/golang_projects/transactions/pkg/transaction"
	"github.com/go-playground/validator"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Router returns the server router instance
func Router(txRepo database.TransactionRepository, v *validator.Validate) http.Handler {
	txHandler := transaction.NewHandler(txRepo, v)

	r := mux.NewRouter()
	r.HandleFunc("/transaction", txHandler.HandleTransaction).Methods(http.MethodPost)

	h := handlers.RecoveryHandler()(r)
	h = handlers.LoggingHandler(os.Stdout, r)
	return h
}
