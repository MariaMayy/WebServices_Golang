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
)

// тут писать SearchServer

type XMLRow struct {
	ID        int      `xml:"id"`
	Age       int      `xml:"age"`
	Gender    string   `xml:"gender"`
	FirstName string   `xml:"first_name"`
	LastName  string   `xml:"last_name"`
	About     string   `xml:"about"`
	XMLName   xml.Name `xml:"row"`
}

type XMLUsers struct {
	XMLName xml.Name `xml:"root"`
	Rows    []XMLRow `xml:"row"`
}

var row XMLUsers
var ResultUsers []User
var MaxLimit = 25

const Token = "Good job!"

func SendErr(w http.ResponseWriter, error string, code int) {
	jsRes, err := json.Marshal(SearchErrorResponse{error})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintln(w, string(jsRes))
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// check AccessToken
	if r.Header.Get("AccessToken") != Token {
		http.Error(w, "Invalid access token", http.StatusUnauthorized)
		return
	}

	File, err := os.Open("dataset.xml") // open file
	defer File.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// read bytes in file
	b, err := ioutil.ReadAll(File)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	xml.Unmarshal(b, &row)
	Query := r.FormValue("query")
	for _, item := range row.Rows {
		if Query != "" {
			bOK := strings.Contains(item.About, Query) || strings.Contains(item.LastName, Query) ||
				strings.Contains(item.FirstName, Query)
			if !bOK {
				continue
			}
		}
		// Added user
		ResultUsers = append(ResultUsers, User{
			Id:     item.ID,
			Age:    item.Age,
			Name:   item.FirstName + " " + item.LastName,
			Gender: item.Gender,
			About:  item.About,
		})
	}

	// Sort
	OrderBy, _ := strconv.Atoi(r.FormValue("order_by"))
	OrderField := r.FormValue("order_field")
	if OrderBy != OrderByAsIs {
		var bOK func(first User, second User) bool
		switch OrderField {
		case "Id":
			bOK = func(first User, second User) bool {
				return first.Id < second.Id
			}
		case "Name":
			bOK = func(first User, second User) bool {
				return first.Name < second.Name
			}
		case "Age":
			bOK = func(first User, second User) bool {
				return first.Age < second.Age
			}
		case "":
			bOK = func(first User, second User) bool {
				return first.Name < second.Name
			}
		default:
			SendErr(w, "ErrorBadOrderField", http.StatusBadRequest)
			return
		}

		sort.Slice(ResultUsers, func(i int, j int) bool {
			return bOK(ResultUsers[i], ResultUsers[j]) && (OrderBy == orderDesc)
		})
	}

	Limit, _ := strconv.Atoi(r.FormValue("limit"))
	Offset, _ := strconv.Atoi(r.FormValue("offset"))

	if Limit > 0 {
		start := Offset
		if start > len(ResultUsers)-1 {
			ResultUsers = []User{}
		} else {
			end := Offset + Limit
			if end > len(ResultUsers) {
				end = len(ResultUsers)
			}
			ResultUsers = ResultUsers[start:end]
		}
	}

	jsRes, err := json.Marshal(ResultUsers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsRes)
}
