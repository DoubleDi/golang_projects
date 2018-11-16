package main

import (
	"fmt"
	"log"
	"reflect"
	"test-task/db"
	"testing"

	"./api"
	"./configuration"
	"./db"
	"./entity/user"

	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	_ "github.com/go-sql-driver/mysql"
)

type Case struct {
	Number int
	Method string
	Path   string
	Query  string
	Status int
	Result interface{}
	Body   interface{}
}

var (
	config = &configuration.Config{
		DBUser: "root",
		DBPass: "",
		DBHost: "localhost",
		DBPort: "3306",
		DBName: "",
	}
)

func PrepareTestData(db *db.DB) {
	qs := []string{
		`drop database if exists testing;`,
		`create database testing`,
		`use testing`,
		`create table users (id integer primary key, age tinyint, sex enum("M","F"));`,
		`create table stats (user_id integer, action enum("like","comment","exit","login"), ts datetime);`,
		`create index action_ts_user_id_idx on stats (action, ts, user_id);`,
	}

	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			log.Panicln(err)
		}
	}
}

func TestApis(t *testing.T) {

	db, err := db.Init(config)
	err = db.Ping()
	if err != nil {
		log.Panicln(err)
	}

	PrepareTestData(db)

	api := &api.API{
		DB: db,
	}
	r := api.InitRouter()

	ts := httptest.NewServer(r)

	cases := []Case{
		Case{
			Number: 1,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-&date2=2018-10-11&limit=10&action=comment",
			Status: http.StatusBadRequest,
		},
		Case{
			Number: 2,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-10&date2=asdfsdf&limit=10&action=comment",
			Status: http.StatusBadRequest,
		},
		Case{
			Number: 3,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-10&date2=2018-10-11&limit=aaaa&action=comment",
			Status: http.StatusBadRequest,
		},
		Case{
			Number: 4,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-10&date2=2018-10-11&limit=10&action=aaaaaaaaaa",
			Status: http.StatusBadRequest,
		},
		Case{
			Number: 5,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-10&date2=2018-10-11",
			Status: http.StatusBadRequest,
		},
		Case{
			Number: 6,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-10&date2=2018-10-11&limit=10",
			Status: http.StatusBadRequest,
		},
		Case{
			Number: 7,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-10&date2=2018-10-11&action=comment",
			Status: http.StatusBadRequest,
		},
		Case{
			Number: 8,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-10&limit=10&action=comment",
			Status: http.StatusBadRequest,
		},
		Case{
			Number: 9,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date2=2018-10-11&limit=10&action=comment",
			Status: http.StatusBadRequest,
		},
		Case{
			Number: 10,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-10&date2=2018-10-11&limit=10&action=comment",
			Result: map[string][]user.UsersByDate{
				"items": []user.UsersByDate{},
			},
			Status: http.StatusOK,
		},
		Case{
			Number: 11,
			Method: http.MethodPost,
			Path:   "/api/users",
			Status: http.StatusCreated,
			Body: map[string]interface{}{
				"id":  1,
				"age": 20,
				"sex": "M",
			},
		},
		Case{
			Number: 12,
			Method: http.MethodPost,
			Path:   "/api/users",
			Status: http.StatusCreated,
			Body: map[string]interface{}{
				"id":  2,
				"age": 20,
				"sex": "F",
			},
		},
		Case{
			Number: 13,
			Method: http.MethodPost,
			Path:   "/api/users",
			Status: http.StatusConflict,
			Body: map[string]interface{}{
				"id":  1,
				"age": 20,
				"sex": "M",
			},
		},
		Case{
			Number: 14,
			Method: http.MethodPost,
			Path:   "/api/users",
			Status: http.StatusBadRequest,
			Body: map[string]interface{}{
				"id":  "asdasd",
				"age": 20,
				"sex": "M",
			},
		},
		Case{
			Number: 15,
			Method: http.MethodPost,
			Path:   "/api/users",
			Status: http.StatusBadRequest,
			Body: map[string]interface{}{
				"id":  1,
				"age": "asdasd",
				"sex": "M",
			},
		},
		Case{
			Number: 16,
			Method: http.MethodPost,
			Path:   "/api/users",
			Status: http.StatusBadRequest,
			Body: map[string]interface{}{
				"id":  1,
				"age": 20,
				"sex": "X",
			},
		},
		Case{
			Number: 17,
			Method: http.MethodPost,
			Path:   "/api/users/stats",
			Status: http.StatusCreated,
			Body: map[string]interface{}{
				"user":   1,
				"action": "comment",
				"ts":     "2017-10-10T14:12:34",
			},
		},
		Case{
			Number: 18,
			Method: http.MethodPost,
			Path:   "/api/users/stats",
			Status: http.StatusBadRequest,
			Body: map[string]interface{}{
				"user":   "asda",
				"action": "comment",
				"ts":     "2017-10-10T14:12:34",
			},
		},
		Case{
			Number: 19,
			Method: http.MethodPost,
			Path:   "/api/users/stats",
			Status: http.StatusBadRequest,
			Body: map[string]interface{}{
				"user":   10,
				"action": "commeasdnt",
				"ts":     "2017-10-10T14:12:34",
			},
		},
		Case{
			Number: 20,
			Method: http.MethodPost,
			Path:   "/api/users/stats",
			Status: http.StatusBadRequest,
			Body: map[string]interface{}{
				"user":   10,
				"action": "comment",
				"ts":     "201:34",
			},
		},
		Case{
			Number: 22,
			Method: http.MethodPost,
			Path:   "/api/users/stats",
			Status: http.StatusCreated,
			Body: map[string]interface{}{
				"user":   1,
				"action": "comment",
				"ts":     "2018-10-10T14:12:34",
			},
		},
		Case{
			Number: 23,
			Method: http.MethodPost,
			Path:   "/api/users/stats",
			Status: http.StatusCreated,
			Body: map[string]interface{}{
				"user":   2,
				"action": "comment",
				"ts":     "2018-10-10T14:12:34",
			},
		},
		Case{
			Number: 24,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-10&date2=2018-10-10&limit=10&action=comment",
			Result: map[string][]user.UsersByDate{
				"items": []user.UsersByDate{
					{
						Date: "2018-10-10",
						Rows: []user.User{
							{
								ID:    1,
								Age:   20,
								Sex:   "M",
								Count: 1,
							},
							{
								ID:    2,
								Age:   20,
								Sex:   "F",
								Count: 1,
							},
						},
					},
					{
						Date: "2017-10-10",
						Rows: []user.User{
							{
								ID:    1,
								Age:   20,
								Sex:   "M",
								Count: 1,
							},
						},
					},
				},
			},
			Status: http.StatusOK,
		},
		Case{
			Number: 25,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-11&date2=2018-10-11&limit=10&action=comment",
			Result: map[string][]user.UsersByDate{
				"items": []user.UsersByDate{
					{
						Date: "2018-10-10",
						Rows: []user.User{
							{
								ID:    1,
								Age:   20,
								Sex:   "M",
								Count: 1,
							},
							{
								ID:    2,
								Age:   20,
								Sex:   "F",
								Count: 1,
							},
						},
					},
				},
			},
			Status: http.StatusOK,
		},
		Case{
			Number: 26,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-10&date2=2018-01-01&limit=10&action=comment",
			Result: map[string][]user.UsersByDate{
				"items": []user.UsersByDate{
					{
						Date: "2017-10-10",
						Rows: []user.User{
							{
								ID:    1,
								Age:   20,
								Sex:   "M",
								Count: 1,
							},
						},
					},
				},
			},
			Status: http.StatusOK,
		},
		Case{
			Number: 27,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-10&date2=2018-10-11&limit=1&action=comment",
			Result: map[string][]user.UsersByDate{
				"items": []user.UsersByDate{
					{
						Date: "2018-10-10",
						Rows: []user.User{
							{
								ID:    1,
								Age:   20,
								Sex:   "M",
								Count: 1,
							},
						},
					},
					{
						Date: "2017-10-10",
						Rows: []user.User{
							{
								ID:    1,
								Age:   20,
								Sex:   "M",
								Count: 1,
							},
						},
					},
				},
			},
			Status: http.StatusOK,
		},
		Case{
			Number: 28,
			Method: http.MethodGet,
			Path:   "/api/users/stats/top",
			Query:  "date1=2017-10-10&date2=2018-10-11&limit=10&action=like",
			Result: map[string][]user.UsersByDate{
				"items": []user.UsersByDate{},
			},
			Status: http.StatusOK,
		},
	}

	for _, item := range cases {
		var (
			err    error
			result map[string][]user.UsersByDate
			req    *http.Request
		)

		caseName := fmt.Sprintf("case %v: [%s] %s %s", item.Number, item.Method, item.Path, item.Query)

		if item.Method == http.MethodGet {
			req, err = http.NewRequest(item.Method, ts.URL+item.Path+"?"+item.Query, nil)
		} else {
			data, err := json.Marshal(item.Body)
			if err != nil {
				panic(err)
			}
			reqBody := bytes.NewReader(data)
			req, err = http.NewRequest(item.Method, ts.URL+item.Path, reqBody)
			req.Header.Add("Content-Type", "application/json")
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("[%s] request error: %v", caseName, err)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if item.Status == 0 {
			item.Status = http.StatusOK
		}

		if resp.StatusCode != item.Status {
			t.Errorf("[%v] expected http status %v, got %v", item.Number, item.Status, resp.StatusCode)
			continue
		}

		if string(body) != "" && resp.StatusCode == http.StatusOK {
			err = json.Unmarshal(body, &result)
			if err != nil {
				t.Errorf("[%s] cant unpack json: %v", caseName, err)
				continue
			}

			if !reflect.DeepEqual(result, item.Result) {
				t.Errorf("[%d] results not match\nGot : %#v\nWant: %#v", item.Number, result, item.Result)
				continue
			}
		}

	}
}
