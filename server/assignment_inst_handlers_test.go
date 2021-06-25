package server

import (
	"bytes"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/assignments"
	"github.com/DAv10195/submit_server/elements/courses"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func getDbForAssInstHandlersTest() (map[string]*users.User, func()) {
	cleanup := db.InitDbForTest()
	testUsers := make(map[string]*users.User)
	if err := users.InitDefaultAdmin(); err != nil {
		panic(err)
	}
	admin, err := users.Get(users.Admin)
	if err != nil {
		panic(err)
	}
	testUsers[users.Admin] = admin
	course, err := courses.NewCourse(1, "course", admin.UserName, true, false)
	if err != nil {
		panic(err)
	}
	user1, err := users.NewUserBuilder(admin.UserName, true).WithUserName("user1").WithPassword("user1").WithRoles(users.StandardUser).
		WithCoursesAsStaff(string(course.Key())).Build()
	if err != nil {
		panic(err)
	}
	testUsers["user1"] = user1
	user2, err := users.NewUserBuilder(admin.UserName, true).WithUserName("user2").WithPassword("user2").WithRoles(users.StandardUser).
		WithCoursesAsStudent(string(course.Key())).Build()
	if err != nil {
		panic(err)
	}
	testUsers["user2"] = user2
	user3, err := users.NewUserBuilder(admin.UserName, true).WithUserName("user3").WithPassword("user3").WithRoles(users.StandardUser).
		WithCoursesAsStudent(string(course.Key())).Build()
	if err != nil {
		panic(err)
	}
	testUsers["user3"] = user3
	_, err = assignments.NewDef(string(course.Key()), time.Now().Add(time.Hour).UTC(), "ass", user1.UserName, true, false)
	if err != nil {
		panic(err)
	}
	return testUsers, cleanup
}

func TestAssInstHandlers(t *testing.T) {
	testUsers, cleanup := getDbForAssInstHandlersTest()
	defer cleanup()
	forUser2Headers, forAssHeaders := make(map[string]string), make(map[string]string)
	year := time.Now().UTC().Year()
	forUser2Headers[submithttp.ForSubmitUser], forAssHeaders[submithttp.ForSubmitAss] = "user2", fmt.Sprintf("1:%d:ass", year)
	router := mux.NewRouter()
	am := NewAuthManager()
	router.Use(contentTypeMiddleware, authenticationMiddleware, am.authorizationMiddleware)
	initAssDefsRouter(router, am)
	initAssInstsRouter(router, am)
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
		name	string
		method	string
		path	string
		status	int
		data	[]byte
		reqUser	*users.User
		headers	map[string]string
	}{
		{
			"test get assignments as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.AssignmentInstances),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test get assignments for user as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.AssignmentInstances),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			forUser2Headers,
		},
		{
			"test get assignments for assignment def as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.AssignmentInstances),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			forAssHeaders,
		},
		{
			"test get assignment as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.AssignmentInstances, year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test update assignment as admin",
			http.MethodPut,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.AssignmentInstances, year),
			http.StatusAccepted,
			[]byte(`{"created_by":"admin","created_on":"2021-06-24T22:37:49.321616Z","updated_by":"admin","updated_on":"2021-06-24T22:37:49.321616Z","user_name":"user2","assignment_def":"1:2021:ass","state":0,"files":{"elements":{"file.txt":{}}},"due_by":"2021-06-24T23:37:49.29625Z","copy":false,"grade":0}`),
			testUsers[users.Admin],
			nil,
		},
		{
			"test submit assignment as admin",
			http.MethodPatch,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.AssignmentInstances, year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test get assignments for assignment def as staff",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.AssignmentInstances),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
			forAssHeaders,
		},
		{
			"test get assignments for assignment def as student",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.AssignmentInstances),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user2"],
			forAssHeaders,
		},
		{
			"test get assignments for assignment def as std user",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.AssignmentInstances),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user1"],
			nil,
		},
		{
			"test get assignment as staff",
			http.MethodGet,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.AssignmentInstances, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user1"],
			nil,
		},
		{
			"test get self assignment",
			http.MethodGet,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.AssignmentInstances, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user2"],
			nil,
		},
		{
			"test get other assignment",
			http.MethodGet,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.AssignmentInstances, year),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user3"],
			nil,
		},
		{
			"test update assignment as staff",
			http.MethodPut,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.AssignmentInstances, year),
			http.StatusAccepted,
			[]byte(`{"created_by":"admin","created_on":"2021-06-24T22:37:49.321616Z","updated_by":"admin","updated_on":"2021-06-24T22:37:49.321616Z","user_name":"user2","assignment_def":"1:2021:ass","state":1,"files":{"elements":{"file.txt":{},"file2.txt":{}}},"due_by":"2021-06-24T23:37:49.29625Z","copy":false,"grade":0}`),
			testUsers["user1"],
			nil,
		},
		{
			"test update assignment as student",
			http.MethodPut,
			fmt.Sprintf("/%s/1/%d/ass/user2", db.AssignmentInstances, year),
			http.StatusAccepted,
			[]byte(`{"created_by":"admin","created_on":"2021-06-24T22:37:49.321616Z","updated_by":"admin","updated_on":"2021-06-24T22:37:49.321616Z","user_name":"user2","assignment_def":"1:2021:ass","state":1,"files":{"elements":{"file.txt":{},"file2.txt":{},"file3.txt":{}}},"due_by":"2021-06-24T23:37:49.29625Z","copy":false,"grade":0}`),
			testUsers["user2"],
			nil,
		},
		{
			"test submit assignment as staff",
			http.MethodPatch,
			fmt.Sprintf("/%s/1/%d/ass/user3", db.AssignmentInstances, year),
			http.StatusForbidden,
			[]byte(""),
			testUsers["user1"],
			nil,
		},
		{
			"test submit assignment as student",
			http.MethodPatch,
			fmt.Sprintf("/%s/1/%d/ass/user3", db.AssignmentInstances, year),
			http.StatusOK,
			[]byte(""),
			testUsers["user3"],
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
