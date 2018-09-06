package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type xmlInput struct {
	Row []User `xml:"row"`
}

const (
	CorrectAccessToken        = "d41d8cd98f00b204e9800998ecf8427e"
	filePath           string = "./dataset.xml"
)

func SearchServer(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("AccessToken")
	if token != CorrectAccessToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	file, _ := os.Open(filePath)

	fileContents, _ := ioutil.ReadAll(file)

	dataSet := &xmlInput{}
	_ = xml.Unmarshal(fileContents, dataSet)
	users := dataSet.Row

	query := r.FormValue("query")
	if query != "" {
		queryUsers := []User{}
		for _, u := range users {
			if strings.Contains(u.FirstName+" "+u.LastName, query) || strings.Contains(u.About, query) {
				queryUsers = append(queryUsers, u)
			}
		}
		users = queryUsers
	}

	orderBy, _ := strconv.Atoi(r.FormValue("order_by"))
	orderField := r.FormValue("order_field")
	if orderField == "" {
		orderField = "Name"
	}
	if orderField != "Id" && orderField != "Name" && orderField != "Age" {
		w.WriteHeader(http.StatusBadRequest)
		jsonErr, _ := json.Marshal(&SearchErrorResponse{Error: ErrorBadOrderField})
		fmt.Fprintf(w, string(jsonErr))
		return
	}
	if orderBy != 0 && orderBy != 1 && orderBy != -1 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if orderField == "Name" {
		if orderBy == OrderByAsc {
			sort.Slice(users, func(i, j int) bool { return users[i].FirstName < users[j].FirstName })
		} else if orderBy == OrderByDesc {
			sort.Slice(users, func(i, j int) bool { return users[i].FirstName > users[j].FirstName })
		}
	} else if orderField == "Id" {
		if orderBy == OrderByAsc {
			sort.Slice(users, func(i, j int) bool { return users[i].Id < users[j].Id })
		} else if orderBy == OrderByDesc {
			sort.Slice(users, func(i, j int) bool { return users[i].Id > users[j].Id })
		}
	} else if orderField == "Age" {
		if orderBy == OrderByAsc {
			sort.Slice(users, func(i, j int) bool { return users[i].Age < users[j].Age })
		} else if orderBy == OrderByDesc {
			sort.Slice(users, func(i, j int) bool { return users[i].Age > users[j].Age })
		}
	}

	var limit int
	var offset int

	offset, _ = strconv.Atoi(r.FormValue("ofset"))

	limit, _ = strconv.Atoi(r.FormValue("limit"))

	if len(users) > limit+offset {
		users = users[offset : limit+offset]
	} else if len(users) < limit+offset {
		if len(users) > offset {
			users = users[offset:]
		} else {
			users = []User{}
		}
	}

	jsonUsers, _ := json.Marshal(&users)

	fmt.Fprintf(w, string(jsonUsers))
}

func Timeout(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 10)
}

func BadJSONBadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "{1}")
	return
}

func BadUserJSON(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{1}")
	return
}

func Forbidden(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	jsonErr, _ := json.Marshal(&SearchErrorResponse{Error: "123"})
	fmt.Fprintf(w, string(jsonErr))
	return
}
