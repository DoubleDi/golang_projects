package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"../entity/user"

	"../entity/stat"

	"github.com/go-sql-driver/mysql"
)

func (api *API) AddUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Printf("Error on loading body %#v %v\n", r, err.Error())
		api.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	user := &user.User{}
	err = json.Unmarshal(body, user)
	if err != nil {
		log.Printf("Invalid JSON %#v %v\n", string(body), err.Error())
		api.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	err = user.Validate()
	if err != nil {
		log.Printf("Invalid user %v\n", err.Error())
		api.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	err = api.DB.AddUser(user)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			log.Printf("User already exists %v\n", mysqlErr.Error())
			api.RespondWithError(w, http.StatusConflict, err)
			return
		} else {
			log.Printf("Error on inserting stat to db %v\n", err.Error())
			api.RespondWithError(w, http.StatusInternalServerError, err)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("User %#v added to DB\n", user)
}

func (api *API) TopUsers(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	limitStr := r.FormValue("limit")
	date1Str := r.FormValue("date1")
	date2Str := r.FormValue("date2")

	if action == "" || limitStr == "" || date1Str == "" || date2Str == "" {
		log.Printf("A required field not defined")
		api.RespondWithError(w, http.StatusBadRequest, fmt.Errorf(
			`a required field not defined action:%v, limit:%v, date1:%v, date2:%v`, action, limitStr, date1Str, date2Str,
		))
		return
	}

	if !stat.IsValidAction(action) {
		log.Printf("Supported actions are %#v\n", stat.AvailibleActions)
		api.RespondWithError(w, http.StatusBadRequest, fmt.Errorf("supported actions are %#v\n", stat.AvailibleActions))
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		log.Printf("Invalid limit %v\n", limitStr)
		api.RespondWithError(w, http.StatusBadRequest, err)
		return
	}
	date1, err := time.Parse(stat.DateLayout, r.FormValue("date1"))
	if err != nil {
		log.Printf("Bad date1 value", err.Error())
		api.RespondWithError(w, http.StatusBadRequest, err)
		return
	}
	date2, err := time.Parse(stat.DateLayout, r.FormValue("date2"))
	if err != nil {
		log.Printf("Bad date2 value", err.Error())
		api.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	if date1.After(date2) {
		log.Printf("date1 > date2")
		api.RespondWithError(w, http.StatusBadRequest, fmt.Errorf("date1 > date2"))
		return
	}

	topUsers, err := api.DB.TopUsers(date1, date2, action, limit)
	if err != nil {
		log.Printf("Can't get top users %v\n", err.Error())
		api.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	topUsersJSON, err := json.Marshal(map[string][]user.UsersByDate{"items": topUsers})
	if err != nil {
		log.Printf("Can't transform struct to JSON %#v %v\n", topUsers, err.Error())
		api.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	log.Printf("Got users %#v\n", topUsers)
	w.WriteHeader(http.StatusOK)
	w.Write(topUsersJSON)
}
