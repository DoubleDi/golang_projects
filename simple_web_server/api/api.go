package api

import (
	"net/http"

	"../db"

	"github.com/gorilla/mux"
)

type API struct {
	DB *db.DB
}

func (api *API) InitRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/api/users", api.AddUser).Methods("POST")
	r.HandleFunc("/api/users/stats", api.AddStat).Methods("POST")
	r.HandleFunc("/api/users/stats/top", api.TopUsers).Methods("GET")

	return r
}

func (api *API) RespondWithError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	w.Write([]byte(err.Error()))
}
