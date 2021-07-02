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

func TestAppealHandlers(t *testing.T) {
	testUsers, cleanup := getDbForAssInstHandlersTest()
	defer cleanup()
	cleanupSess := session.InitSessionForTest()
	defer cleanupSess()
	forCourseHeaders, forAssHeaders, forAssUser2Headers, forAssUser3Headers, stateCloseHeaders, stateOpenHeaders := make(map[string]string), make(map[string]string), make(map[string]string), make(map[string]string), make(map[string]string), make(map[string]string)
	year := time.Now().UTC().Year()
	forCourseHeaders[submithttp.ForSubmitCourse] = fmt.Sprintf("1:%d", year)
	forAssHeaders[submithttp.ForSubmitAss], forAssUser2Headers[submithttp.ForSubmitAss], forAssUser3Headers[submithttp.ForSubmitAss] = fmt.Sprintf("1:%d:ass", year), fmt.Sprintf("1:%d:ass:user2", year), fmt.Sprintf("1:%d:ass:user3", year)
	stateCloseHeaders[submithttp.SubmitState], stateOpenHeaders[submithttp.SubmitState] = submithttp.AppealStateClosed, submithttp.AppealStateOpen
	router := mux.NewRouter()
	am := NewAuthManager()
	router.Use(contentTypeMiddleware, authenticationMiddleware, am.authorizationMiddleware)
	initAssDefsRouter(router, am)
	initAppealsRouter(router, am)
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
	assUser2, err := assignments.GetInstance(fmt.Sprintf("1:%d:ass:user2", year))
	if err != nil {
		panic(err)
	}
	assUser2.State = assignments.Graded
	assUser3, err := assignments.GetInstance(fmt.Sprintf("1:%d:ass:user3", year))
	if err != nil {
		panic(err)
	}
	assUser3.State = assignments.Graded
	if err := db.Update(testUsers[users.Admin].UserName, assUser2, assUser3); err != nil {
		panic(err)
	}
	testCases := []struct{
		name	string
		method	string
		path	string
		status	int
		data	[]byte
		reqUser	*users.User
		headers	map[string]string
	}{
		{
			"test create appeal as admin",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.Appeals),
			http.StatusAccepted,
			[]byte(""),
			testUsers[users.Admin],
			forAssUser2Headers,
		},
		{
			"test create appeal as staff",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.Appeals),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user1"],
			forAssUser3Headers,
		},
		{
			"test create appeal as student",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.Appeals),
			http.StatusAccepted,
			[]byte(""),
			testUsers["user3"],
			forAssUser3Headers,
		},
		{
			"test get appeals as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Appeals),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test get appeals as std user",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Appeals),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user1"],
			nil,
		},
		{
			"test get appeals as admin for course",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Appeals),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			forCourseHeaders,
		},
		{
			"test get appeals as staff for course",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Appeals),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
			forCourseHeaders,
		},
		{
			"test get appeals as student for course",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Appeals),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user2"],
			forCourseHeaders,
		},
		{
			"test get appeals as admin for assignment",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Appeals),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			forAssHeaders,
		},
		{
			"test get appeals as staff for assignment",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Appeals),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
			forAssHeaders,
		},
		{
			"test get appeals as student for assignment",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Appeals),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user2"],
			forAssHeaders,
		},
		{
			"test get appeal as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.Appeals, year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test get appeal as staff",
			http.MethodGet,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.Appeals, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
			nil,
		},
		{
			"test get self appeal",
			http.MethodGet,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.Appeals, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user2"],
			nil,
		},
		{
			"test get other user appeal",
			http.MethodGet,
			fmt.Sprintf("/%s/1/%d/ass/user3", db.Appeals, year),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user2"],
			nil,
		},
		{
			"test update appeal state as admin",
			http.MethodPatch,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.Appeals, year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			stateCloseHeaders,
		},
		{
			"test update self appeal state",
			http.MethodPatch,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.Appeals, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user2"],
			stateOpenHeaders,
		},
		{
			"test update appeal state as staff",
			http.MethodPatch,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.Appeals, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
			stateCloseHeaders,
		},
		{
			"test update other user appeal state",
			http.MethodPatch,
			fmt.Sprintf("/%s/1/%d/ass/user3", db.Appeals, year),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user2"],
			stateCloseHeaders,
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
