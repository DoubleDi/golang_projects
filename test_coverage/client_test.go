package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestDummy(t *testing.T) {
	// t.Errorf("TODO")
}

func TestWork(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      5,
		Offset:     0,
		Query:      "la",
		OrderField: "Age",
		OrderBy:    1,
	})

	if err != nil {
		t.Errorf("work process error: %v", err)
	}
}

func TestLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	resp, err := client.FindUsers(SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "la",
		OrderField: "Age",
		OrderBy:    -1,
	})

	if err != nil {
		t.Errorf("work process error: %v", err)
	}

	if len(resp.Users) != 2 {
		t.Error("limit not working")
	}
}

func TestUndefLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	resp1, err := client.FindUsers(SearchRequest{
		Offset:     0,
		Query:      "la",
		OrderField: "Id",
		OrderBy:    1,
	})

	if err != nil {
		t.Errorf("work process error: %v", err)
	}

	resp2, err := client.FindUsers(SearchRequest{
		Limit:      1000,
		Offset:     0,
		Query:      "la",
		OrderField: "Name",
		OrderBy:    0,
	})

	if err != nil {
		t.Errorf("work process error: %v", err)
	}

	if len(resp1.Users) != len(resp2.Users) || len(resp1.Users) != 25 {
		t.Error("limit undefined not working")
	}
}

func TestOffset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	resp1, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "la",
		OrderField: "Id",
		OrderBy:    -1,
	})

	if err != nil {
		t.Errorf("work process error: %v", err)
	}

	resp2, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     1,
		Query:      "la",
		OrderField: "Name",
		OrderBy:    0,
	})

	if err != nil {
		t.Errorf("work process error: %v", err)
	}

	if !reflect.DeepEqual(resp1.Users[1], resp2.Users[0]) {
		t.Error("offset not working")
	}
}

func TestBigOffset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	resp, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     1000,
		Query:      "la",
		OrderField: "Name",
		OrderBy:    0,
	})

	if err != nil {
		t.Errorf("work process error: %v", err)
	}

	if len(resp.Users) > 0 {
		t.Error("offset not working")
	}
}

func TestOrderBy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	resp1, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Query:      "la",
		OrderField: "Name",
		OrderBy:    0,
	})

	if err != nil {
		t.Errorf("work process error: %v", err)
	}

	resp2, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "la",
		OrderField: "Name",
		OrderBy:    1,
	})

	if err != nil {
		t.Errorf("work process error: %v", err)
	}

	resp3, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "la",
		OrderField: "Name",
		OrderBy:    -1,
	})

	if err != nil {
		t.Errorf("work process error: %v", err)
	}

	reversedResp3 := resp3.Users
	for i, j := 0, len(reversedResp3)-1; i < len(reversedResp3)/2; i, j = i+1, j-1 {
		reversedResp3[i], reversedResp3[j] = reversedResp3[j], reversedResp3[i]
	}
	if !reflect.DeepEqual(resp1.Users, resp2.Users) &&
		!reflect.DeepEqual(resp1.Users, resp3.Users) &&
		!reflect.DeepEqual(resp2.Users, resp3.Users) &&
		reflect.DeepEqual(resp2.Users, reversedResp3) {
		t.Error("sort not working")
	}
}

func TestOrderByField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	resp1, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "la",
		OrderField: "",
		OrderBy:    0,
	})

	if err != nil {
		t.Errorf("work process error: %v", err)
	}

	resp2, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "la",
		OrderField: "Id",
		OrderBy:    0,
	})

	if err != nil {
		t.Errorf("work process error: %v", err)
	}

	resp3, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "la",
		OrderField: "Age",
		OrderBy:    0,
	})

	if err != nil {
		t.Errorf("work process error: %v", err)
	}

	reversedResp3 := resp3.Users
	for i, j := 0, len(reversedResp3)-1; i < len(reversedResp3)/2; i, j = i+1, j-1 {
		reversedResp3[i], reversedResp3[j] = reversedResp3[j], reversedResp3[i]
	}
	if !reflect.DeepEqual(resp1.Users, resp2.Users) &&
		!reflect.DeepEqual(resp1.Users, resp3.Users) &&
		!reflect.DeepEqual(resp2.Users, resp3.Users) {
		t.Error("sort fields not working")
	}
}

func TestWrongOrderField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	orderField := "asdfkasldfj"
	_, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     1000,
		Query:      "la",
		OrderField: orderField,
		OrderBy:    0,
	})

	if err.Error() != "OrderFeld asdfkasldfj invalid" {
		t.Error("wrong order field not working")
	}
}

func TestNegativeLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      -1,
		Offset:     1000,
		Query:      "la",
		OrderField: "Name",
		OrderBy:    0,
	})

	if err.Error() != "limit must be > 0" {
		t.Error("limit < 0 error")
	}
}

func TestNegativeOffset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     -1,
		Query:      "la",
		OrderField: "Name",
		OrderBy:    0,
	})

	if err.Error() != "offset must be > 0" {
		t.Error("offset < 0 error")
	}
}

func TestBadAccessToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "cf8427e",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     1,
		Query:      "la",
		OrderField: "Name",
		OrderBy:    0,
	})

	if err.Error() != "Bad AccessToken" {
		t.Error("bad token error")
	}
}

func TestInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "la",
		OrderField: "Name",
		OrderBy:    2,
	})

	if err.Error() != "SearchServer fatal error" {
		t.Error("internal server bad error")
	}
}

func Test(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Timeout))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "la",
		OrderField: "Name",
		OrderBy:    2,
	})

	if !strings.Contains(err.Error(), "timeout") {
		t.Error("itimeout bad test error")
	}
}

func TestBadJSONBadRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(BadJSONBadRequest))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "la",
		OrderField: "Name",
		OrderBy:    2,
	})

	if !strings.Contains(err.Error(), "cant unpack error json") {
		t.Error("itimeout bad test error")
	}
}

func TestBadUserJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(BadUserJSON))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "la",
		OrderField: "",
		OrderBy:    2,
	})

	if !strings.Contains(err.Error(), "cant unpack result json") {
		t.Error("itimeout bad test error")
	}
}

func TestSearchAbout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	resp, _ := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "sit velit exercitation non aliqua",
		OrderField: "Name",
		OrderBy:    0,
	})

	if len(resp.Users) == 0 {
		t.Error("about field search error")
	}
}

func TestBadUrl(t *testing.T) {
	_ = httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         "asdasd",
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "sit velit exercitation non aliqua",
		OrderField: "Name",
		OrderBy:    0,
	})

	if err == nil {
		t.Error("bad url error fail ")
	}
}

func TestForbidden(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Forbidden))

	client := SearchClient{
		AccessToken: "d41d8cd98f00b204e9800998ecf8427e",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "sit velit exercitation non aliqua",
		OrderField: "Name",
		OrderBy:    0,
	})

	fmt.Println(err)
	if err == nil {
		t.Error("bad url error fail ")
	}
}

func TestMain(t *testing.T) {
	main()
}
