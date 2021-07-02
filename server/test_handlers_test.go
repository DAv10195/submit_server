package server

import (
	"bytes"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/session"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTestHandlers(t *testing.T) {
	testUsers, cleanup := getDbForAssInstHandlersTest()
	defer cleanup()
	cleanupSess := session.InitSessionForTest()
	defer cleanupSess()
	forAssHeaders := make(map[string]string)
	year := time.Now().UTC().Year()
	forAssHeaders[submithttp.ForSubmitAss] = fmt.Sprintf("1:%d:ass", year)
	forAssUser1Headers := make(map[string]string)
	forAssUser1Headers[submithttp.ForSubmitAss] = fmt.Sprintf("1:%d:ass", year)
	forAssUser1Headers[submithttp.ForSubmitUser] = "user1"
	forAssUser2Headers := make(map[string]string)
	forAssUser2Headers[submithttp.ForSubmitAss] = fmt.Sprintf("1:%d:ass", year)
	forAssUser2Headers[submithttp.ForSubmitUser] = "user2"
	router := mux.NewRouter()
	am := NewAuthManager()
	router.Use(contentTypeMiddleware, authenticationMiddleware, am.authorizationMiddleware)
	initAssDefsRouter(router, am)
	initTestsRouter(router, am)
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("/%s/1/%d/ass", db.AssignmentDefinitions, year), bytes.NewBuffer([]byte("")))
	if err != nil {
		panic(err)
	}
	pwd, err := db.Decrypt(testUsers[users.Admin].Password)
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(testUsers[users.Admin].UserName, pwd)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		panic("error publishing assignment for test")
	}
	testCases := []struct{
		name			string
		method			string
		path			string
		status			int
		data			[]byte
		reqUser			*users.User
		headers			map[string]string
	}{
		{
			"test get tests as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Tests),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test get tests as admin for ass",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Tests),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			forAssHeaders,
		},
		{
			"test get tests as admin for ass user1",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Tests),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			forAssUser1Headers,
		},
		{
			"test create test as admin",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.Tests),
			http.StatusAccepted,
			[]byte(fmt.Sprintf(`{"name":"test1","assignment_def":"1:%d:ass","runs_on":0}`, year)),
			testUsers[users.Admin],
			nil,
		},
		{
			"test get test as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/1/%d/ass/test1", db.Tests, year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test update test as admin",
			http.MethodPut,
			fmt.Sprintf("/%s/1/%d/ass/test1", db.Tests, year),
			http.StatusAccepted,
			[]byte(fmt.Sprintf(`{"created_by":"admin","created_on":"2021-06-24T22:37:49.321616Z","updated_by":"admin","updated_on":"2021-06-24T22:37:49.321616Z","name":"test1","assignment_def":"1:%d:ass","runs_on":1,"files":{"elements":{"file.txt":{}}},"state":0}`, year)),
			testUsers[users.Admin],
			nil,
		},
		{
			"test update test status as admin",
			http.MethodPatch,
			fmt.Sprintf("/%s/1/%d/ass/test1", db.Tests, year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test delete test as admin",
			http.MethodDelete,
			fmt.Sprintf("/%s/1/%d/ass/test1", db.Tests, year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test get tests as std user",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Tests),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user1"],
			nil,
		},
		{
			"test get tests as staff for ass",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Tests),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
			forAssHeaders,
		},
		{
			"test get tests for ass and user as staff",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Tests),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
			forAssUser1Headers,
		},
		{
			"test create test as staff",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.Tests),
			http.StatusAccepted,
			[]byte(fmt.Sprintf(`{"name":"test1","assignment_def":"1:%d:ass","runs_on":0,"files":{"elements":{}}}`, year)),
			testUsers["user1"],
			nil,
		},
		{
			"test get test as staff",
			http.MethodGet,
			fmt.Sprintf("/%s/1/%d/ass/test1", db.Tests, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
			nil,
		},
		{
			"test update test as staff",
			http.MethodPut,
			fmt.Sprintf("/%s/1/%d/ass/test1", db.Tests, year),
			http.StatusAccepted,
			[]byte(fmt.Sprintf(`{"name":"test1","assignment_def":"1:%d:ass","runs_on":1,"files":{"elements":{"file.txt":{}}}}`, year)),
			testUsers["user1"],
			nil,
		},
		{
			"test update test status as staff",
			http.MethodPatch,
			fmt.Sprintf("/%s/1/%d/ass/test1", db.Tests, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
			nil,
		},
		{
			"test delete test as staff",
			http.MethodDelete,
			fmt.Sprintf("/%s/1/%d/ass/test1", db.Tests, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
			nil,
		},
		{
			"test get tests as student for ass",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Tests),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user2"],
			forAssHeaders,
		},
		{
			"test get tests for ass and user as student",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Tests),
			http.StatusOK,
			[]byte(""),
			testUsers["user2"],
			forAssUser2Headers,
		},
		{
			"test create on submit test as student",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.Tests),
			http.StatusForbidden,
			[]byte(fmt.Sprintf(`{"name":"test1","assignment_def":"1:%d:ass","runs_on":0,"files":{"elements":{}}}`, year)),
			testUsers["user2"],
			nil,
		},
		{
			"test create on demand test as student",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.Tests),
			http.StatusAccepted,
			[]byte(fmt.Sprintf(`{"name":"test1","assignment_def":"1:%d:ass","runs_on":1,"files":{"elements":{}}}`, year)),
			testUsers["user2"],
			nil,
		},
		{
			"test get test as student",
			http.MethodGet,
			fmt.Sprintf("/%s/1/%d/ass/test1", db.Tests, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user2"],
			nil,
		},
		{
			"test update test as student",
			http.MethodPut,
			fmt.Sprintf("/%s/1/%d/ass/test1", db.Tests, year),
			http.StatusAccepted,
			[]byte(fmt.Sprintf(`{"name":"test1","assignment_def":"1:%d:ass","runs_on":1,"files":{"elements":{"file.txt":{}}}}`, year)),
			testUsers["user2"],
			nil,
		},
		{
			"test update test status as student 1",
			http.MethodPatch,
			fmt.Sprintf("/%s/1/%d/ass/test1", db.Tests, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user2"],
			nil,
		},
		{
			"test update test status as student 2",
			http.MethodPatch,
			fmt.Sprintf("/%s/1/%d/ass/test1", db.Tests, year),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user2"],
			nil,
		},
		{
			"test delete test as student",
			http.MethodDelete,
			fmt.Sprintf("/%s/1/%d/ass/test1", db.Tests, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user2"],
			nil,
		},
	}
	for _, testCase := range testCases {
		var testCaseErr error
		if !t.Run(testCase.name, func (t *testing.T) {
			r, err := http.NewRequest(testCase.method, testCase.path, bytes.NewBuffer(testCase.data))
			if err != nil {
				testCaseErr = fmt.Errorf("error creating http request for test case [ %s ]: %v", testCase.name, err)
				t.FailNow()
			}
			password, err := db.Decrypt(testCase.reqUser.Password)
			if err != nil {
				testCaseErr = fmt.Errorf("error decrypting password for http request in test case [ %s ]: %v", testCase.name, err)
				t.FailNow()
			}
			r.SetBasicAuth(testCase.reqUser.UserName, password)
			if testCase.headers != nil {
				for k, v := range testCase.headers {
					r.Header.Set(k, v)
				}
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			if w.Code != testCase.status {
				testCaseErr = fmt.Errorf("test case [ %s ] produced status code %d instead of the expected %d status code", testCase.name, w.Code, testCase.status)
				t.FailNow()
			}
		}) {
			t.Logf("error in test case [ %s ]: %v", testCase.name, testCaseErr)
		}
	}
}
