package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestUnpackFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Wow, error, bro:c", http.StatusBadRequest)
		}))
	defer ts.Close()
	sClient := SearchClient{AccessToken, ts.URL}

	sc := SearchRequest{
		Limit:      5,
		Offset:     1,
		Query:      "",
		OrderField: "",
		OrderBy:    OrderByAsc,
	}
	_, err := sClient.FindUsers(sc)
	if err == nil {
		t.Errorf("Empty error, bro")
	} else if !strings.Contains(err.Error(), "cant unpack error json") {
		t.Errorf("Invalid error: %v", err.Error())
	}

}

func TestOffsetFailDown(t *testing.T) {
	ts := startTServer(AccessToken)
	defer ts.Close()

	sc := SearchRequest{
		Limit:      8,
		Offset:     -5,
		Query:      "",
		OrderField: "",
		OrderBy:    OrderByAsc,
	}
	_, err := ts.search.FindUsers(sc)
	if err == nil {
		t.Error("offset must be > 0")
	} else if err.Error() != "offset must be > 0" {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestBadOrderField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			SendErr(w, ErrorBadOrderField, http.StatusBadRequest)
		}))
	defer ts.Close()
	sClient := SearchClient{AccessToken, ts.URL}

	sc := SearchRequest{
		Limit:      5,
		Offset:     1,
		Query:      "",
		OrderField: "Bad",
		OrderBy:    OrderByAsc,
	}
	_, err := sClient.FindUsers(sc)
	if err == nil {
		t.Errorf("Empty error, bro")
	} else if !strings.Contains(err.Error(), "OrderFeld Bad invalid") {
		t.Errorf("Invalid error: %v", err.Error())
	}

}

func TestUnknownError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			SendErr(w, "Unknown error((", http.StatusBadRequest)
		}))
	defer ts.Close()
	sClient := SearchClient{AccessToken, ts.URL}

	sc := SearchRequest{
		Limit:      5,
		Offset:     1,
		Query:      "",
		OrderField: "",
		OrderBy:    OrderByAsc,
	}
	_, err := sClient.FindUsers(sc)
	if err == nil {
		t.Errorf("Empty error, bro")
	} else if !strings.Contains(err.Error(), "unknown bad request error:") {
		t.Errorf("Invalid error: %v", err.Error())
	}

}

func TestTimeoutError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
		}))
	defer ts.Close()
	sClient := SearchClient{AccessToken, ts.URL}

	sc := SearchRequest{
		Limit:      5,
		Offset:     1,
		Query:      "",
		OrderField: "",
		OrderBy:    OrderByAsc,
	}
	_, err := sClient.FindUsers(sc)
	if err == nil {
		t.Errorf("Empty error, bro")
	} else if !strings.Contains(err.Error(), "timeout for") {
		t.Errorf("Invalid error: %v", err.Error())
	}

}

func TestUnknownErr(t *testing.T) {
	sClient := SearchClient{AccessToken, "BAD"}

	sc := SearchRequest{
		Limit:      5,
		Offset:     1,
		Query:      "",
		OrderField: "",
		OrderBy:    OrderByAsc,
	}
	_, err := sClient.FindUsers(sc)
	if err == nil {
		t.Errorf("Empty error, bro")
	} else if !strings.Contains(err.Error(), "unknown error") {
		t.Errorf("Invalid error: %v", err.Error())
	}

}

func TestFatalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Wow, fatal error!", http.StatusInternalServerError)
		}))
	defer ts.Close()
	sClient := SearchClient{AccessToken, ts.URL}

	sc := SearchRequest{
		Limit:      5,
		Offset:     1,
		Query:      "",
		OrderField: "",
		OrderBy:    OrderByAsc,
	}
	_, err := sClient.FindUsers(sc)
	if err == nil {
		t.Errorf("Empty error, bro")
	} else if err.Error() != "SearchServer fatal error" {
		t.Errorf("Invalid error: %v", err.Error())
	}

}
