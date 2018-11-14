package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	yaml "gopkg.in/yaml.v2"

	_ "github.com/go-sql-driver/mysql"
)

type Handler struct {
	DB *sql.DB
}

type User struct {
	ID    int    `json:"id"`
	Age   int    `json:"age"`
	Sex   string `json:"sex"`
	Count int    `json:"count"`
}

type Config struct {
	DBUser string `yaml:"DBUser"`
	DBPass string `yaml:"DBPass"`
	DBHost string `yaml:"DBHost"`
	DBPort string `yaml:"DBPort"`
	DBName string `yaml:"DBName"`
}

var (
	actionList     = []string{"like", "comment", "exit", "login"}
	config         = &Config{}
	datetimeLayout = "2006-01-02T15:04:05"
	dateLayout     = "2006-01-02"
)

func IsValidAction(checkAction string) bool {
	for _, action := range actionList {
		if checkAction == action {
			return true
		}
	}

	return false
}

func ErrorRecovery(w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		log.Println("recovered", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.(string)))
	}
}

func (h *Handler) AddUser(w http.ResponseWriter, r *http.Request) {
	defer ErrorRecovery(w, r)

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Panicf("Error on loading body %#v %v\n", r, err.Error())
	}

	user := &User{}
	err = json.Unmarshal(body, user)
	if err != nil {
		log.Panicf("Invalid JSON %#v %v\n", string(body), err.Error())
	}

	_, err = h.DB.Exec("insert into users values (?, ?, ?)", user.ID, user.Age, user.Sex)
	if err != nil {
		log.Panicln("Something is wrong with DB", err.Error())
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("User %#v added to DB\n", user)
}

func (h *Handler) AddStat(w http.ResponseWriter, r *http.Request) {
	defer ErrorRecovery(w, r)

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Panicf("Error on loading body %#v %v\n", r, err.Error())
	}

	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Panicf("Invalid JSON %#v %v\n", string(body), err.Error())
	}

	userID := int(data["user"].(float64))
	ts, err := time.Parse(datetimeLayout, data["ts"].(string))
	if err != nil {
		log.Panicln("Bad ts value", err.Error())
	}
	action := data["action"].(string)
	if !IsValidAction(action) {
		log.Panicf("Supported actions are %#v\n", actionList)
	}

	_, err = h.DB.Exec("insert into stats values (?, ?, ?)", userID, action, ts)
	if err != nil {
		log.Panicln("Something is wrong with DB", err.Error())
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("Stat UserID: %v Action: %v TS: %v added to DB\n", userID, action, ts)
}

func (h *Handler) TopStats(w http.ResponseWriter, r *http.Request) {
	defer ErrorRecovery(w, r)

	action := r.FormValue("action")
	if !IsValidAction(action) {
		log.Panicf("Supported actions are %#v\n", actionList)
	}
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		limit = 100
	}
	date1, err := time.Parse(dateLayout, r.FormValue("date1"))
	if err != nil {
		log.Panicln("Bad date1 value", err.Error())
	}
	date2, err := time.Parse(dateLayout, r.FormValue("date2"))
	if err != nil {
		log.Panicln("Bad date2 value", err.Error())
	}

	rows, err := h.DB.Query(`
		select id, age, sex, date(ts) as date, count(*) as count from users 
		join stats on (users.id=stats.user_id) 
		where date(stats.ts) >= ? and date(stats.ts) <= ? and action = ? 
		group by stats.user_id, date(ts) 
		order by count desc, date desc limit ?;
	`, date1, date2, action, limit)
	if err != nil {
		log.Panicln("Something is wrong with DB", err.Error())
	}

	res := []map[string]interface{}{}
	userDate := ""
	userRows := []User{}
	for rows.Next() {
		var (
			date string
			user User
		)
		err = rows.Scan(&user.ID, &user.Age, &user.Sex, &date, &user.Count)
		if err != nil {
			log.Panicln("Something is wrong with scanning values", err.Error())
		}
		if userDate == "" {
			userDate = date
		}

		if date != userDate {
			res = append(res, map[string]interface{}{
				"date": userDate,
				"rows": userRows,
			})
			userDate = ""
			userRows = []User{}
		}

		userRows = append(userRows, user)
	}

	rows.Close()
	if userDate != "" {
		res = append(res, map[string]interface{}{
			"date": userDate,
			"rows": userRows,
		})
	}

	jsonRes, err := json.Marshal(res)
	if err != nil {
		log.Panicf("Can't transform struct to JSON %#v %v\n", res, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonRes)
}

func init() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	configFile, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Panicln("Error on open config file", err.Error())
	}
	yaml.Unmarshal(configFile, &config)
}

func main() {
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8",
		config.DBUser, config.DBPass, config.DBHost, config.DBPort, config.DBName,
	)
	log.Println("Connecting to db", dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Panicln("Error on opening DB connection", err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Panicln("Error on pinging DB connection", err.Error())
	}
	defer db.Close()

	handler := &Handler{
		DB: db,
	}
	r := mux.NewRouter()
	r.HandleFunc("/api/users", handler.AddUser).Methods("POST")
	r.HandleFunc("/api/users/stats", handler.AddStat).Methods("POST")
	r.HandleFunc("/api/users/stats/top", handler.TopStats).Methods("GET")

	http.ListenAndServe(":8080", r)
}
