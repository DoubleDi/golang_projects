package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"../entity/stat"

	"github.com/go-sql-driver/mysql"
)

func (api *API) AddStat(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Printf("Error on loading body %#v %v\n", r, err.Error())
		api.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	stat := &stat.Stat{}
	err = json.Unmarshal(body, &stat)

	if err != nil {
		log.Printf("Invalid JSON %#v %v\n", string(body), err.Error())
		api.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	err = stat.Validate()
	if err != nil {
		log.Printf("Invalid stat %v\n", err.Error())
		api.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	err = api.DB.AddStat(stat)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 { // row already exists
			log.Printf("Stat already exists %v\n", mysqlErr.Error())
			api.RespondWithError(w, http.StatusConflict, mysqlErr)
		} else {
			log.Printf("Error on inserting stat to db %v\n", err.Error())
			api.RespondWithError(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("Stat %#v\n", stat)
}
