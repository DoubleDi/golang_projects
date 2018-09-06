package main

import "encoding/json"
import "net/http"
import "strconv"
import "fmt"
import "context"

type Response struct {
	Error    string      `json:"error"`
	Response interface{} `json:"response,omitempty"`
}

func MakeError(text string) string {
	errResp := &Response{Error: text}
	errStr, err := json.Marshal(errResp)

	if err != nil {
		panic("JSON ERROR")
	}

	return string(errStr)
}

func MakeResponse(error string, response interface{}) string {
	resp := &Response{Error: error, Response: response}
	respStr, err := json.Marshal(resp)

	if err != nil {
		panic("JSON ERROR")
	}

	return string(respStr)
}

func ValidateMyApi(w http.ResponseWriter, r *http.Request) (MyApi, error) {
	str := MyApi{}
	return str, nil
}

func ValidateProfileParams(w http.ResponseWriter, r *http.Request) (ProfileParams, error) {
	str := ProfileParams{}
	// Field Login
	// Ordinary parameter name
	login := r.FormValue("login")

	// Required validation
	if login == "" {
		return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("login must me not empty")}
	}

	// Type cast to string
	loginValue := login

	str.Login = loginValue

	return str, nil
}

func ValidateCreateParams(w http.ResponseWriter, r *http.Request) (CreateParams, error) {
	str := CreateParams{}
	// Field Login
	// Ordinary parameter name
	login := r.FormValue("login")

	// Required validation
	if login == "" {
		return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("login must me not empty")}
	}

	// Type cast to string
	loginValue := login

	// Check min
	loginMin := 10
	if loginMin > len(loginValue) {
		return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("login len must be >= 10")}
	}

	str.Login = loginValue

	// Field Name
	// Custom parameter name
	name := r.FormValue("full_name")

	// Type cast to string
	nameValue := name

	str.Name = nameValue

	// Field Status
	// Ordinary parameter name
	status := r.FormValue("status")

	// Type cast to string
	statusValue := status

	// Check default
	if statusValue == "" {
		statusValue = "user"
	}

	// Check enum
	statusEnum := []string{"user", "moderator", "admin"}
	enumFlag := false
	for _, v := range statusEnum {
		if v == statusValue {
			enumFlag = true
		}
	}

	if !enumFlag {
		return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("status must be one of [user, moderator, admin]")}
	}

	str.Status = statusValue

	// Field Age
	// Ordinary parameter name
	age := r.FormValue("age")

	// Type cast to int
	ageValue, err := strconv.Atoi(age)
	if err != nil {
		return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("age must be int")}
	}

	// Check min
	ageMin := 0
	if ageMin > ageValue {
		return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("age must be >= 0")}
	}

	// Check max
	ageMax := 128
	if ageMax < ageValue {
		return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("age must be <= 128")}
	}

	str.Age = ageValue

	return str, nil
}

func ValidateUser(w http.ResponseWriter, r *http.Request) (User, error) {
	str := User{}
	return str, nil
}

func ValidateNewUser(w http.ResponseWriter, r *http.Request) (NewUser, error) {
	str := NewUser{}
	return str, nil
}

func ValidateApiError(w http.ResponseWriter, r *http.Request) (ApiError, error) {
	str := ApiError{}
	return str, nil
}

func (str *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
	// params validation
	params, err := ValidateProfileParams(w, r)
	if err != nil {
		err := err.(*ApiError)
		w.WriteHeader(err.HTTPStatus)
		fmt.Fprintln(w, MakeError(err.Error()))
		return
	}

	ctx, _ := context.WithCancel(context.Background())
	res, err := str.Profile(ctx, params)
	switch err.(type) {
	case ApiError:
		err := err.(ApiError)
		w.WriteHeader(err.HTTPStatus)
		fmt.Fprintf(w, MakeError(err.Error()))
	case error:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, MakeError(err.Error()))
	default:
		fmt.Fprintf(w, MakeResponse("", res))
	}
}

func (str *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	// check method POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Fprintln(w, MakeError("bad method"))
		return
	}

	// check auth
	auth := r.Header.Get("X-Auth")
	if auth == "" {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, MakeError("unauthorized"))
		return
	}

	// params validation
	params, err := ValidateCreateParams(w, r)
	if err != nil {
		err := err.(*ApiError)
		w.WriteHeader(err.HTTPStatus)
		fmt.Fprintln(w, MakeError(err.Error()))
		return
	}

	ctx, _ := context.WithCancel(context.Background())
	res, err := str.Create(ctx, params)
	switch err.(type) {
	case ApiError:
		err := err.(ApiError)
		w.WriteHeader(err.HTTPStatus)
		fmt.Fprintf(w, MakeError(err.Error()))
	case error:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, MakeError(err.Error()))
	default:
		fmt.Fprintf(w, MakeResponse("", res))
	}
}

func (str *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Path
	switch query {
	case "/user/profile":
		str.handlerProfile(w, r)
	case "/user/create":
		str.handlerCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, MakeError(`unknown method`))
	}
}
