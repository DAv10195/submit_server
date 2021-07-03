package server

import (
	"bytes"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/assignments"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/session"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMessageHandlers(t *testing.T) {
	testUsers, cleanup := getDbForAssInstHandlersTest()
	defer cleanup()
	cleanupSess := session.InitSessionForTest()
	defer cleanupSess()
	router := mux.NewRouter()
	am := NewAuthManager()
	router.Use(contentTypeMiddleware, authenticationMiddleware, am.authorizationMiddleware)
	initAssDefsRouter(router, am)
	initAppealsRouter(router, am)
	initTestsRouter(router, am)
	initMessagesRouter(router, am)
	year := time.Now().UTC().Year()
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("/%s/1/%d/ass", db.AssignmentDefinitions, year), bytes.NewBuffer([]byte("")))
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth("user1", "user1")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		panic("error publishing assignment for test")
	}
	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("/%s/", db.Tests), bytes.NewBuffer([]byte(fmt.Sprintf(`{"assignment_def":"1:%d:ass","runs_on":1,"name":"test"}`, year))))
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth("user2", "user2")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusAccepted {
		panic("error creating test for test")
	}
	assUser2, err := assignments.GetInstance(fmt.Sprintf("1:%d:ass:user2", year))
	if err != nil {
		panic(err)
	}
	assUser2.State = assignments.Graded
	if err := db.Update(testUsers[users.Admin].UserName, assUser2); err != nil {
		panic(err)
	}
	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("/%s/", db.Appeals), bytes.NewBuffer([]byte("")))
	if err != nil {
		panic(err)
	}
	req.Header.Set(submithttp.ForSubmitAss, fmt.Sprintf("1:%d:ass:user2", year))
	req.SetBasicAuth("user2", "user2")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusAccepted {
		panic("error creating appeal for test")
	}
	testCases := []struct{
		name			string
		method			string
		path			string
		status			int
		data			[]byte
		reqUser			*users.User
	}{
		{
			"test get all message boxes as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Messages),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
		},
		{
			"test get all message boxes as std_user",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Messages),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user1"],
		},
		{
			"post message for user as admin",
			http.MethodPost,
			fmt.Sprintf("/%s/%s/user1", db.Messages, db.Users),
			http.StatusAccepted,
			[]byte(`{"text":"hello"}`),
			testUsers[users.Admin],
		},
		{
			"post message for user as std_user",
			http.MethodPost,
			fmt.Sprintf("/%s/%s/user1", db.Messages, db.Users),
			http.StatusAccepted,
			[]byte(`{"text":"hello"}`),
			testUsers["user2"],
		},
		{
			"get user messages as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/%s/user1", db.Messages, db.Users),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
		},
		{
			"get user messages - self request",
			http.MethodGet,
			fmt.Sprintf("/%s/%s/user1", db.Messages, db.Users),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
		},
		{
			"get user messages - other non admin user request",
			http.MethodGet,
			fmt.Sprintf("/%s/%s/user1", db.Messages, db.Users),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user2"],
		},
		{
			"get appeal messages as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/%s/1/%d/ass/user2", db.Messages, db.Appeals, year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
		},
		{
			"get appeal messages as course staff",
			http.MethodGet,
			fmt.Sprintf("/%s/%s/1/%d/ass/user2", db.Messages, db.Appeals, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
		},
		{
			"get appeal messages as appeal user",
			http.MethodGet,
			fmt.Sprintf("/%s/%s/1/%d/ass/user2", db.Messages, db.Appeals, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user2"],
		},
		{
			"get appeal messages as other user",
			http.MethodGet,
			fmt.Sprintf("/%s/%s/1/%d/ass/user2", db.Messages, db.Appeals, year),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user3"],
		},
		{
			"post appeal message as admin",
			http.MethodPost,
			fmt.Sprintf("/%s/%s/1/%d/ass/user2", db.Messages, db.Appeals, year),
			http.StatusAccepted,
			[]byte(`{"text":"hello"}`),
			testUsers[users.Admin],
		},
		{
			"post appeal message as course staff",
			http.MethodPost,
			fmt.Sprintf("/%s/%s/1/%d/ass/user2", db.Messages, db.Appeals, year),
			http.StatusAccepted,
			[]byte(`{"text":"hello"}`),
			testUsers["user1"],
		},
		{
			"post appeal message as appeal student",
			http.MethodPost,
			fmt.Sprintf("/%s/%s/1/%d/ass/user2", db.Messages, db.Appeals, year),
			http.StatusAccepted,
			[]byte(`{"text":"hello"}`),
			testUsers["user2"],
		},
		{
			"post appeal message as other student",
			http.MethodPost,
			fmt.Sprintf("/%s/%s/1/%d/ass/user2", db.Messages, db.Appeals, year),
			http.StatusForbidden,
			[]byte(`{"text":"hello"}`),
			testUsers["user3"],
		},
		{
			"get test messages as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/%s/1/%d/ass/test", db.Messages, db.Tests, year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
		},
		{
			"get test messages as course staff",
			http.MethodGet,
			fmt.Sprintf("/%s/%s/1/%d/ass/test", db.Messages, db.Tests, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
		},
		{
			"get test messages as test user",
			http.MethodGet,
			fmt.Sprintf("/%s/%s/1/%d/ass/test", db.Messages, db.Tests, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user2"],
		},
		{
			"get test messages as other user",
			http.MethodGet,
			fmt.Sprintf("/%s/%s/1/%d/ass/test", db.Messages, db.Tests, year),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user3"],
		},
		{
			"post test message as admin",
			http.MethodPost,
			fmt.Sprintf("/%s/%s/1/%d/ass/test", db.Messages, db.Tests, year),
			http.StatusAccepted,
			[]byte(`{"text":"hello"}`),
			testUsers[users.Admin],
		},
		{
			"post test message as course staff",
			http.MethodPost,
			fmt.Sprintf("/%s/%s/1/%d/ass/test", db.Messages, db.Tests, year),
			http.StatusAccepted,
			[]byte(`{"text":"hello"}`),
			testUsers["user1"],
		},
		{
			"post test message as test user",
			http.MethodPost,
			fmt.Sprintf("/%s/%s/1/%d/ass/test", db.Messages, db.Tests, year),
			http.StatusAccepted,
			[]byte(`{"text":"hello"}`),
			testUsers["user2"],
		},
		{
			"post test message as other user",
			http.MethodPost,
			fmt.Sprintf("/%s/%s/1/%d/ass/test", db.Messages, db.Tests, year),
			http.StatusForbidden,
			[]byte(`{"text":"hello"}`),
			testUsers["user3"],
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
