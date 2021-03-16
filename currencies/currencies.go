package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type Currency struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type Handler struct {
	l          sync.RWMutex
	currencies []Currency
}

func (h *Handler) GetCurrencies(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	h.l.RLock()
	defer h.l.RUnlock()
	log.Println("Retreiving currencies")
	if err := json.NewEncoder(w).Encode(h.currencies); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) AddCurrency(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	c := &Currency{}
	if err := json.NewDecoder(r.Body).Decode(c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("adding currency %#v\n", c)
	h.l.Lock()
	h.currencies = append(h.currencies, *c)
	h.l.Unlock()
}

func main() {
	r := mux.NewRouter()
	h := &Handler{}
	r.HandleFunc("/currency", h.GetCurrencies).Methods(http.MethodGet)
	r.HandleFunc("/currency", h.AddCurrency).Methods(http.MethodPost)
	http.Handle("/", r)
	log.Println("Starting server")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
