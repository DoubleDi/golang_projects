package wallet

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/DoubleDi/golang_projects/billionaire/pkg/database"
)

type BalanceHandler struct {
	repo database.WalletHistoryRepository
}

func NewHandler(repo database.WalletHistoryRepository) *BalanceHandler {
	return &BalanceHandler{
		repo: repo,
	}
}

type HandleBalanceArgs database.Balance

func (h *BalanceHandler) AddBalance(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	args := &HandleBalanceArgs{}
	if err := json.Unmarshal(data, args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !args.Amount.IsPositive() {
		http.Error(w, "amount less than zero", http.StatusBadRequest)
		return
	}

	b := database.Balance(*args)
	b.DateTime = toStartOfHour(b.DateTime)
	log.Printf("Saving balance %+v", b)
	if err := h.repo.AddBalance(r.Context(), &b); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	w.Write([]byte(`{"status":"ok"}`))
	return
}

func (h *BalanceHandler) GetBalances(w http.ResponseWriter, r *http.Request) {
	var (
		query = r.URL.Query()
		to    = time.Now()
		from  = to.AddDate(0, 0, -1)
		err   error
	)

	fromStr := query.Get("startDatetime")
	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			http.Error(w, "invalid startDatetime value", http.StatusBadRequest)
			return
		}
	}
	toStr := query.Get("endDatetime")
	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			http.Error(w, "invalid endDatetime value", http.StatusBadRequest)
			return
		}
	}

	if from.After(to) {
		http.Error(w, "startDatetime after endDatetime", http.StatusBadRequest)
		return
	}
	from = toStartOfHour(from)
	to = toStartOfHour(to)

	log.Printf("Getting balances %v-%v", from, to)
	balances, err := h.repo.GetBalances(r.Context(), from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	if err := json.NewEncoder(w).Encode(balances); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
