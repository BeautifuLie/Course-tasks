package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"

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
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("AccessToken") != "token" {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
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
	// query := r.FormValue("query")
	orderField := r.FormValue("order_field")

	if orderField == "" {
		orderField = "Name"
	}

	if orderField != "Id" && orderField != "Age" && orderField != "Name" {
		resp := SearchErrorResponse{Error: "ErrorBadOrderField"}
		js, _ := json.Marshal(resp)
		http.Error(w, "", http.StatusBadRequest)
		w.Write([]byte(js))
		return
	}

	orderBy, err := strconv.Atoi(r.FormValue("order_by"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	SortBy(orderBy, orderDesc, orderAsc, orderField, users)
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
	t.Run("5", func(t *testing.T) {

		req := SearchRequest{
			Limit:      0,
			Offset:     0,
			Query:      "",
			OrderField: "Idd",
			OrderBy:    0,
		}
		_, err := client.FindUsers(req)
		assert.Error(t, err)

	})
	t.Run("6", func(t *testing.T) {

		req := SearchRequest{
			Limit:      25,
			Offset:     30,
			Query:      "",
			OrderField: "",
			OrderBy:    0,
		}
		res, err := client.FindUsers(req)
		require.NoError(t, err)
		if len(res.Users) != 5 {
			t.Errorf("offset error: %v", len(res.Users))
		}

	})
}

func SortBy(orderBy, OrderByDesc, OrderByAsc int, orderField string, users []User) {
	if orderBy == OrderByDesc {
		switch orderField {
		case "Id":
			sort.Slice(users, func(i, j int) bool {
				return users[i].Id < users[j].Id
			})
		case "Age":
			sort.Slice(users, func(i, j int) bool {
				return users[i].Age < users[j].Age
			})
		case "Name":
			sort.Slice(users, func(i, j int) bool {
				return users[i].Name < users[j].Name
			})
		}
	} else if orderBy == OrderByAsc {
		switch orderField {
		case "Id":
			sort.Slice(users, func(i, j int) bool {
				return users[i].Id > users[j].Id
			})
		case "Age":
			sort.Slice(users, func(i, j int) bool {
				return users[i].Age > users[j].Age
			})
		case "Name":
			sort.Slice(users, func(i, j int) bool {
				return users[i].Name > users[j].Name
			})
		}
	}

}

func testErrorServer(t *testing.T, f func(w http.ResponseWriter, r *http.Request)) {
	ts := httptest.NewServer(http.HandlerFunc(f))
	defer ts.Close()

	cli := &SearchClient{"token", ts.URL}
	req := SearchRequest{}

	_, err := cli.FindUsers(req)
	require.Error(t, err)
}

func TestBadJsonServerError(t *testing.T) {
	testErrorServer(t, BadJsonError)
	testErrorServer(t, UnknownBadRequestServer)
	testErrorServer(t, UnknownError)
	testErrorServer(t, InternalErrorServer)
	testErrorServer(t, Unauthorized)
	testErrorServer(t, BadUserJsonResponseServer)
}

func BadJsonError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusBadRequest)
}
func UnknownBadRequestServer(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusBadRequest)
	w.Write([]byte("{}"))
}

func UnknownError(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "unknown://1234", http.StatusFound)
}

func InternalErrorServer(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusInternalServerError)
}
func Unauthorized(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusUnauthorized)
}

func BadUserJsonResponseServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(""))
}

func TimeoutedServer(w http.ResponseWriter, r *http.Request) {

	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		return
	}
}
func TestTimeoutServerr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(TimeoutedServer))
	defer ts.Close()

	cli := &SearchClient{"token", ts.URL}
	req := SearchRequest{}

	_, err := cli.FindUsers(req)

	require.Error(t, err)
}
