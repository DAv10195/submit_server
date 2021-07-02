package server

import (
	"bytes"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/assignments"
	"github.com/DAv10195/submit_server/elements/courses"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/session"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func getDbForAssDefHandlersTest() (map[string]*users.User, map[string]*courses.Course, map[string]*assignments.AssignmentDef, func()) {
	testUsers, testCourses, cleanup := getDbForCoursesHandlersTest()
	testAsses := make(map[string]*assignments.AssignmentDef)
	ass1, err := assignments.NewDef(string(testCourses["course1"].Key()), time.Now().Add(time.Hour).UTC(), "ass1", db.System, true, false)
	if err != nil {
		panic(err)
	}
	testAsses["ass1"] = ass1
	ass2, err := assignments.NewDef(string(testCourses["course2"].Key()), time.Now().Add(time.Hour).UTC(), "ass2", db.System, true, false)
	if err != nil {
		panic(err)
	}
	testAsses["ass2"] = ass2
	return testUsers, testCourses, testAsses, cleanup
}

func TestAssDefsHandlers(t *testing.T) {
	testUsers, testCourses, testAsses, cleanup := getDbForAssDefHandlersTest()
	defer cleanup()
	cleanupSess := session.InitSessionForTest()
	defer cleanupSess()
	forCourse1HeaderMap, forCourse2HeaderMap := make(map[string]string), make(map[string]string)
	forCourse1HeaderMap[submithttp.ForSubmitCourse] = string(testCourses["course1"].Key())
	forCourse2HeaderMap[submithttp.ForSubmitCourse] = string(testCourses["course2"].Key())
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
			fmt.Sprintf("/%s/", db.AssignmentDefinitions),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test get assignments for course as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.AssignmentDefinitions),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			forCourse1HeaderMap,
		},
		{
			"test create assignment def as admin",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.AssignmentDefinitions),
			http.StatusAccepted,
			[]byte(fmt.Sprintf("{\"course\":\"%s\",\"name\":\"ass3\",\"due_by\":\"2030-06-21T00:00:00.00000Z\"}", string(testCourses["course1"].Key()))),
			testUsers[users.Admin],
			nil,
		},
		{
			"test get assignment def as admin",
			http.MethodGet,
			fmt.Sprintf("/%s/%d/%d/ass3", db.AssignmentDefinitions, testCourses["course1"].Number, testCourses["course1"].Year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test update assignment def as admin",
			http.MethodPut,
			fmt.Sprintf("/%s/%d/%d/ass3", db.AssignmentDefinitions, testCourses["course1"].Number, testCourses["course1"].Year),
			http.StatusAccepted,
			[]byte(fmt.Sprintf("{\"course\":\"%s\",\"name\":\"ass3\",\"due_by\":\"2030-06-21T00:00:00.00000Z\",\"files\":{\"elements\":{\"file.txt\":{}}}}", string(testCourses["course1"].Key()))),
			testUsers[users.Admin],
			nil,
		},
		{
			"test publish assignment def as admin",
			http.MethodPatch,
			fmt.Sprintf("/%s/%d/%d/ass3", db.AssignmentDefinitions, testCourses["course1"].Number, testCourses["course1"].Year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test delete assignment def as admin",
			http.MethodDelete,
			fmt.Sprintf("/%s/%d/%d/ass3", db.AssignmentDefinitions, testCourses["course1"].Number, testCourses["course1"].Year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test get assignments as staff/student",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.AssignmentDefinitions),
			http.StatusForbidden,
			[]byte(""),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test get assignments for course as staff",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.AssignmentDefinitions),
			http.StatusOK,
			[]byte(""),
			testUsers[users.StandardUser],
			forCourse2HeaderMap,
		},
		{
			"test create assignment def as staff",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.AssignmentDefinitions),
			http.StatusAccepted,
			[]byte(fmt.Sprintf("{\"course\":\"%s\",\"name\":\"ass3\",\"due_by\":\"2030-06-21T00:00:00.00000Z\"}", string(testCourses["course2"].Key()))),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test get assignment def as staff",
			http.MethodGet,
			fmt.Sprintf("/%s/%d/%d/ass3", db.AssignmentDefinitions, testCourses["course2"].Number, testCourses["course2"].Year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test update assignment def as staff",
			http.MethodPut,
			fmt.Sprintf("/%s/%d/%d/ass3", db.AssignmentDefinitions, testCourses["course2"].Number, testCourses["course2"].Year),
			http.StatusAccepted,
			[]byte(fmt.Sprintf("{\"course\":\"%s\",\"name\":\"ass3\",\"due_by\":\"2030-06-21T00:00:00.00000Z\",\"files\":{\"elements\":{\"file.txt\":{}}}}", string(testCourses["course2"].Key()))),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test publish assignment def as staff",
			http.MethodPatch,
			fmt.Sprintf("/%s/%d/%d/ass3", db.AssignmentDefinitions, testCourses["course2"].Number, testCourses["course2"].Year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test delete assignment def as staff",
			http.MethodDelete,
			fmt.Sprintf("/%s/%d/%d/ass3", db.AssignmentDefinitions, testCourses["course2"].Number, testCourses["course2"].Year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test get assignments for course as student",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.AssignmentDefinitions),
			http.StatusForbidden,
			[]byte(""),
			testUsers[users.StandardUser],
			forCourse1HeaderMap,
		},
		{
			"test create assignment def as student",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.AssignmentDefinitions),
			http.StatusForbidden,
			[]byte(fmt.Sprintf("{\"course\":\"%s\",\"name\":\"ass3\",\"due_by\":\"2030-06-21T00:00:00.00000Z\"}", string(testCourses["course1"].Key()))),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test get assignment def as student",
			http.MethodGet,
			fmt.Sprintf("/%s/%d/%d/%s", db.AssignmentDefinitions, testCourses["course1"].Number, testCourses["course1"].Year, testAsses["ass1"].Name),
			http.StatusForbidden,
			[]byte(""),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test update assignment def as student",
			http.MethodPut,
			fmt.Sprintf("/%s/%d/%d/%s", db.AssignmentDefinitions, testCourses["course1"].Number, testCourses["course1"].Year, testAsses["ass1"].Name),
			http.StatusForbidden,
			[]byte(fmt.Sprintf("{\"course\":\"%s\",\"name\":\"ass1\",\"due_by\":\"2030-06-21T00:00:00.00000Z\",\"files\":{\"elements\":{\"file.txt\":{}}}}", string(testCourses["course2"].Key()))),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test publish assignment def as student",
			http.MethodPatch,
			fmt.Sprintf("/%s/%d/%d/%s", db.AssignmentDefinitions, testCourses["course1"].Number, testCourses["course1"].Year, testAsses["ass1"].Name),
			http.StatusForbidden,
			[]byte(""),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test delete assignment def as student",
			http.MethodDelete,
			fmt.Sprintf("/%s/%d/%d/%s", db.AssignmentDefinitions, testCourses["course1"].Number, testCourses["course1"].Year, testAsses["ass1"].Name),
			http.StatusForbidden,
			[]byte(""),
			testUsers[users.StandardUser],
			nil,
		},
	}
	router := mux.NewRouter()
	am := NewAuthManager()
	router.Use(contentTypeMiddleware, authenticationMiddleware, am.authorizationMiddleware)
	initAssDefsRouter(router, am)
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
