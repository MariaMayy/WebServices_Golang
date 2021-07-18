package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// тут писать код тестов
const AccessToken = "Good job!"

type TServer struct {
	server *httptest.Server
	search SearchClient
}

func startTServer(token string) TServer {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	sc := SearchClient{token, ts.URL}

	return TServer{ts, sc}
}

func (ts *TServer) Close() {
	ts.server.Close()
}

func TestLimitFailDown(t *testing.T) {
	ts := startTServer(AccessToken)
	defer ts.Close()

	sc := SearchRequest{
		Limit:      -4,
		Offset:     1,
		Query:      "",
		OrderField: "",
		OrderBy:    OrderByAsc,
	}
	_, err := ts.search.FindUsers(sc)
	if err == nil {
		t.Error("limit must be > 0")
	} else if err.Error() != "limit must be > 0" {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestLimitFailUp(t *testing.T) {
	ts := startTServer(AccessToken)
	defer ts.Close()

	sc := SearchRequest{
		Limit:      100,
		Offset:     1,
		Query:      "",
		OrderField: "",
		OrderBy:    OrderByAsc,
	}
	res, _ := ts.search.FindUsers(sc)
	if len(res.Users) != 25 {
		t.Errorf("Invalid number of users: %d", len(res.Users))
	}
}

func TestBadToken(t *testing.T) {
	ts := startTServer(AccessToken + "invalid")
	defer ts.Close()

	sc := SearchRequest{
		Limit:      5,
		Offset:     1,
		Query:      "",
		OrderField: "",
		OrderBy:    OrderByAsc,
	}
	_, err := ts.search.FindUsers(sc)
	if err == nil {
		t.Errorf("Empty error, bro")
	} else if err.Error() != "Bad AccessToken" {
		t.Errorf("Invalid error: %v", err.Error())
	}

}
