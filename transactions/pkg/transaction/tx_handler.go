package transaction

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/shopspring/decimal"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"

	"github.com/DoubleDi/golang_projects/transactions/pkg/database"
	"github.com/go-playground/validator"
)

// TransactionHandler encapsulates the transaction web handlers
type TransactionHandler struct {
	repo      database.TransactionRepository
	validator *validator.Validate
}

// NewHandler returns a new TransactionHandler
func NewHandler(repo database.TransactionRepository, validator *validator.Validate) *TransactionHandler {
	return &TransactionHandler{
		repo:      repo,
		validator: validator,
	}
}

// HandleTransaction handles a new transaction and writes it to db
func (h *TransactionHandler) HandleTransaction(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if !validateSourceType(r.Header.Get("Source-Type")) {
		http.Error(w, "invalid 'Source-Type' header", http.StatusBadRequest)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	args := &HandleTransactionArgs{}
	if err := json.Unmarshal(data, args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.validator.Struct(args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if args.Amount.LessThan(decimal.Zero) {
		http.Error(w, "amount less than zero", http.StatusBadRequest)
		return
	}
	t := database.Transaction{
		ID:     args.TransactionID,
		State:  args.State,
		Amount: args.Amount,
	}
	log.Printf("Saving transaction %+v", t)
	if err := h.repo.SaveTransaction(&t); err != nil {
		pgerr := pgx.PgError{}
		if errors.As(err, &pgerr) && pgerr.Code == pgerrcode.UniqueViolation {
			http.Error(w, "transaction with id '"+t.ID+"' already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
	return
}
