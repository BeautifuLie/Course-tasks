package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type XMLUser struct {
	ID        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}
type XMLUsers struct {
	Users []XMLUser `xml:"row"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {

	file, err := os.Open("dataset.xml")
	if err != nil {
		http.Error(w, "Can't open XML-file", http.StatusInternalServerError)
	}
	defer file.Close()
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Can't read XML-file", http.StatusInternalServerError)
	}
	var xmlUsers XMLUsers
	err = xml.Unmarshal(bytes, &xmlUsers)
	if err != nil {
		http.Error(w, "Error unmarshall XML-file", http.StatusInternalServerError)
	}
	var users []User
	for _, xuser := range xmlUsers.Users {
		user := User{
			Id:     xuser.ID,
			Name:   xuser.FirstName + "" + xuser.LastName,
			Age:    xuser.Age,
			About:  xuser.About,
			Gender: xuser.Gender,
		}
		users = append(users, user)
	}
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	max := offset + limit
	if max > len(users) {
		max = len(users)
	}

	js, err := json.Marshal(users[offset:max])

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(js)
	// query:=r.FormValue("query")

}

func TestSearchServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{"token", ts.URL}

	t.Run("1", func(t *testing.T) {
		req := SearchRequest{
			Limit:      -1,
			Offset:     0,
			Query:      "",
			OrderField: "",
			OrderBy:    0,
		}
		_, err := client.FindUsers(req)

		var limitError = errors.New("limit must be > 0")
		require.Equal(t, err, limitError)
	})
	t.Run("2", func(t *testing.T) {
		req := SearchRequest{
			Limit:      0,
			Offset:     -1,
			Query:      "",
			OrderField: "",
			OrderBy:    0,
		}
		_, err := client.FindUsers(req)

		var offsetError = errors.New("offset must be > 0")
		require.Equal(t, err, offsetError)
	})
	t.Run("3", func(t *testing.T) {
		req := SearchRequest{
			Limit:      26,
			Offset:     0,
			Query:      "",
			OrderField: "",
			OrderBy:    0,
		}
		res, _ := client.FindUsers(req)

		// assert.NotEqual(t, req1.Limit, usersLength)

		assert.NotEqual(t, req.Limit, len(res.Users))

	})
	t.Run("4", func(t *testing.T) {
		limit := 9
		req := SearchRequest{
			Limit:      limit,
			Offset:     0,
			Query:      "",
			OrderField: "",
			OrderBy:    0,
		}
		res, _ := client.FindUsers(req)
		assert.Equal(t, limit, len(res.Users))

	})
}
