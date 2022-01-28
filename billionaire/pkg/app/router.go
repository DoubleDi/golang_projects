package app

import (
	"net/http"
	"os"

	"github.com/DoubleDi/golang_projects/billionaire/pkg/database"
	"github.com/DoubleDi/golang_projects/billionaire/pkg/wallet"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func Router(repo database.WalletHistoryRepository) http.Handler {
	walletHandler := wallet.NewHandler(repo)

	r := mux.NewRouter()
	r.HandleFunc("/balance", walletHandler.AddBalance).Methods(http.MethodPost)
	r.HandleFunc("/balance", walletHandler.GetBalances).Methods(http.MethodGet)

	h := handlers.RecoveryHandler()(r)
	h = handlers.LoggingHandler(os.Stdout, r)
	return h
}
