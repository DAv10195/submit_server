package server

import (
	"fmt"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/users"
	//"github.com/DAv10195/submit_server/util/containers"
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
	server = initServer()
	path := os.TempDir()
	if err := db.InitDB(path); err != nil {
		t.Fatal(err)
	}
	testAddPathToMap()
	testAddRegex()

	resp, err := http.Get("http://localhost:8080/users/")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatal(err)
	}
	resp, err = http.Get("http://localhost:8080/users/admin")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatal(err)
	}
}

func negativeTest(t *testing.T){
	server = initServer()
	path := os.TempDir()
	if err := db.InitDB(path); err != nil {
		t.Fatal(err)
	}
	testAddPathToMap()
	testAddRegex()

	resp, err := http.Get("http://localhost:8080/users/")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatal(err)
	}
	resp, err = http.Get("http://localhost:8080/users/admin")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatal(err)
	}
	//req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/", nil)
	//req.SetBasicAuth("user", "password")
	//
	//resp, _ = http.DefaultClient.Do(req)
	//respBytes, _ := ioutil.ReadAll(resp.Body)
	//var bla interface{}
	//_ := json.Unmarshal(respBytes, bla)
}



func TestInit(t *testing.T) {
	negativeTest(t)
	positiveTest(t)
}