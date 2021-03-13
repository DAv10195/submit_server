package server

import (
	"fmt"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"regexp"
	"testing"
)

var port = 8080
var am *authManager
var server *http.Server

func initServer() *http.Server {
	baseRouter := mux.NewRouter()
	baseRouter.Use(contentTypeMiddleware)
	baseRouter.Use(authenticationMiddleware)
	am = &authManager{authMap: make(map[string]authorizationFunc)}
	baseRouter.Use(am.authorizationMiddleware)
	initUsersRouter(baseRouter, am)
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      baseRouter,
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
	}
}

func testAddPathToMap() {
	am.addPathToMap("/users/", func(user *users.User) bool{
		if user.UserName == "admin" {
			return true
		}
		return false
	})
}

func testAddRegex() {
	regex, _ := regexp.Compile("/users/.")
	am.addRegex(regex, func(user *users.User) bool{
		if user.UserName == "admin" {
			return true
		}
		return false
	})
}


func positiveTest(t *testing.T){
	//init the server and the db.
	server = initServer()
	go server.ListenAndServe()
	path := os.TempDir()
	if err := db.InitDB(path); err != nil {
		t.Fatal(err)
	}
	// using the admin user , lets set a path and a authFunc.
	testAddPathToMap()
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/users/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("admin", "admin")
	resp, err2 := http.DefaultClient.Do(req)
	if err2 != nil {
		t.Fatal(err2)
	}
	if resp.StatusCode != http.StatusOK{
		t.Fatal("test failed.")
	}
	// using the admin , lets set a regex and a authFunc.
	testAddRegex()
	req, err = http.NewRequest(http.MethodGet, "http://localhost:8080/users/admin", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("admin", "admin")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK{
		t.Fatal("test failed.")
	}

	//respBytes, _ := ioutil.ReadAll(resp.Body)
	//var bla interface{}
	//_ := json.Unmarshal(respBytes, bla)
}

func negativeTest(t *testing.T){
	//init the server and the db.
	server = initServer()
	go server.ListenAndServe()
	path := os.TempDir()
	if err := db.InitDB(path); err != nil {
		t.Fatal(err)
	}
	//register new user with all permissions. but the username is not admin.

	testAddPathToMap()
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/users/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("nikita", "nikita")
	resp, err2 := http.DefaultClient.Do(req)
	if err2 != nil {
		t.Fatal(err2)
	}
	testAddRegex()
	if resp.StatusCode != http.StatusForbidden{
		t.Fatal("test failed.")
	}
	req, err = http.NewRequest(http.MethodGet, "http://localhost:8080/users/admin", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("nikita", "admin")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusForbidden{
		t.Fatal("test failed.")
	}
}


func TestInit(t *testing.T) {
	negativeTest(t)
	positiveTest(t)
}