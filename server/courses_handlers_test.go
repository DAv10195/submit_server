package server

import (
	"bytes"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/courses"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func getDbForCoursesHandlersTest() (map[string]*users.User, map[string]*courses.Course, func()) {
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
	secretary, err := users.NewUserBuilder(db.System, true).WithUserName(users.Secretary).WithPassword(users.Secretary).WithRoles(users.Secretary).Build()
	if err != nil {
		panic(err)
	}
	testUsers[users.Secretary] = secretary
	testCourses := make(map[string]*courses.Course)
	course1, err := courses.NewCourse(1, "course1", secretary.UserName, true, false)
	if err != nil {
		panic(err)
	}
	testCourses["course1"] = course1
	course2, err := courses.NewCourse(2, "course2", secretary.UserName, true, false)
	if err != nil {
		panic(err)
	}
	testCourses["course2"] = course2
	stdUser, err := users.NewUserBuilder(db.System, true).WithUserName(users.StandardUser).WithPassword(users.StandardUser).WithRoles(users.StandardUser).
		WithCoursesAsStudent(string(course1.Key())).WithCoursesAsStaff(string(course2.Key())).Build()
	if err != nil {
		panic(err)
	}
	testUsers[users.StandardUser] = stdUser
	return testUsers, testCourses, cleanup
}

func TestCoursesHandlers(t *testing.T) {
	testUsers, testCourses, cleanup := getDbForCoursesHandlersTest()
	defer cleanup()
	stdUserHeader, adminHeaderMap := make(map[string]string), make(map[string]string)
	stdUserHeader[submithttp.ForSubmitUser], adminHeaderMap[submithttp.ForSubmitUser] = users.StandardUser, users.Admin
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
			"test get all courses with admin",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Courses),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test get all courses with secretary",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Courses),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Secretary],
			nil,
		},
		{
			"test get all courses with std_user",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Courses),
			http.StatusForbidden,
			[]byte(""),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test create course with admin",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.Courses),
			http.StatusAccepted,
			[]byte("{\"name\":\"course3\",\"number\":3}"),
			testUsers[users.Admin],
			nil,
		},
		{
			"test create course with secretary",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.Courses),
			http.StatusAccepted,
			[]byte("{\"name\":\"course4\",\"number\":4}"),
			testUsers[users.Secretary],
			nil,
		},
		{
			"test create course with std_user",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.Courses),
			http.StatusForbidden,
			[]byte("{\"name\":\"course5\",\"number\":5}"),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test get course with admin",
			http.MethodGet,
			fmt.Sprintf("/%s/%d/%d", db.Courses, testCourses["course1"].Number, testCourses["course1"].Year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test get course with secretary",
			http.MethodGet,
			fmt.Sprintf("/%s/%d/%d", db.Courses, testCourses["course1"].Number, testCourses["course1"].Year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Secretary],
			nil,
		},
		{
			"test get course with std_user as student",
			http.MethodGet,
			fmt.Sprintf("/%s/%d/%d", db.Courses, testCourses["course1"].Number, testCourses["course1"].Year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test get course with std_user as staff",
			http.MethodGet,
			fmt.Sprintf("/%s/%d/%d", db.Courses, testCourses["course2"].Number, testCourses["course2"].Year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test update course with std_user as staff",
			http.MethodPut,
			fmt.Sprintf("/%s/%d/%d", db.Courses, testCourses["course2"].Number, testCourses["course2"].Year),
			http.StatusAccepted,
			[]byte(fmt.Sprintf("{\"name\":\"%s\",\"number\":%d,\"year\":%d}", testCourses["course2"].Name, testCourses["course2"].Number, testCourses["course2"].Year)),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test update course with std_user as student",
			http.MethodPut,
			fmt.Sprintf("/%s/%d/%d", db.Courses, testCourses["course1"].Number, testCourses["course1"].Year),
			http.StatusForbidden,
			[]byte(fmt.Sprintf("{\"name\":\"%s\",\"number\":%d,\"year\":%d}", testCourses["course1"].Name, testCourses["course1"].Number, testCourses["course1"].Year)),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test update course with admin",
			http.MethodPut,
			fmt.Sprintf("/%s/%d/%d", db.Courses, testCourses["course1"].Number, testCourses["course1"].Year),
			http.StatusAccepted,
			[]byte(fmt.Sprintf("{\"name\":\"%s\",\"number\":%d,\"year\":%d}", testCourses["course1"].Name, testCourses["course1"].Number, testCourses["course1"].Year)),
			testUsers[users.Admin],
			nil,
		},
		{
			"test update course with secretary",
			http.MethodPut,
			fmt.Sprintf("/%s/%d/%d", db.Courses, testCourses["course1"].Number, testCourses["course1"].Year),
			http.StatusAccepted,
			[]byte(fmt.Sprintf("{\"name\":\"%s\",\"number\":%d,\"year\":%d}", testCourses["course1"].Name, testCourses["course1"].Number, testCourses["course1"].Year)),
			testUsers[users.Secretary],
			nil,
		},
		{
			"test delete course with std_user",
			http.MethodDelete,
			fmt.Sprintf("/%s/%d/%d", db.Courses, testCourses["course1"].Number, testCourses["course1"].Year),
			http.StatusForbidden,
			[]byte(""),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test delete course with secretary",
			http.MethodDelete,
			fmt.Sprintf("/%s/%d/%d", db.Courses, testCourses["course1"].Number, testCourses["course1"].Year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Secretary],
			nil,
		},
		{
			"test delete course with admin",
			http.MethodDelete,
			fmt.Sprintf("/%s/%d/%d", db.Courses, testCourses["course2"].Number, testCourses["course2"].Year),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			nil,
		},
		{
			"test get non existent course",
			http.MethodGet,
			fmt.Sprintf("/%s/%d/%d", db.Courses, 1234, testCourses["course2"].Year),
			http.StatusNotFound,
			[]byte(""),
			testUsers[users.StandardUser],
			nil,
		},
		{
			"test get std_user courses",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Courses),
			http.StatusOK,
			[]byte(""),
			testUsers[users.StandardUser],
			stdUserHeader,
		},
		{
			"test get std_user courses with admin",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Courses),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
			stdUserHeader,
		},
		{
			"test get admin courses with stdUser",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Courses),
			http.StatusForbidden,
			[]byte(""),
			testUsers[users.StandardUser],
			adminHeaderMap,
		},
	}
	router := mux.NewRouter()
	am := NewAuthManager()
	router.Use(contentTypeMiddleware, authenticationMiddleware, am.authorizationMiddleware)
	initCoursesRouter(router, am)
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
