package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func initDbForUsersHandlersTest() (map[string]*users.User, func()) {
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
	stdUser, err := users.NewUserBuilder(db.System, true).WithUserName(users.StandardUser).WithPassword(users.StandardUser).WithRoles(users.StandardUser).Build()
	if err != nil {
		panic(err)
	}
	testUsers[users.StandardUser] = stdUser
	return testUsers, cleanup
}

func TestUsersHandlers(t *testing.T) {
	testUsers, cleanup := initDbForUsersHandlersTest()
	testUsersBytes := make(map[string][]byte)
	for k, v := range testUsers {
		userBytes, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		testUsersBytes[k] = userBytes
	}
	defer cleanup()
	testCases := []struct{
		name	string
		method	string
		path	string
		status	int
		data	[]byte
		reqUser	*users.User
	}{
		{
			"test get all users with admin",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Users),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
		},
		{
			"test get all users with secretary",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Users),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Secretary],
		},
		{
			"test get all users with std_user",
			http.MethodGet,
			fmt.Sprintf("/%s/", db.Users),
			http.StatusForbidden,
			[]byte(""),
			testUsers[users.StandardUser],
		},
		{
			"test register users with admin",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.Users),
			http.StatusAccepted,
			[]byte("{\"users\":[{\"user_name\":\"test1\",\"password\":\"test\",\"roles\":{\"elements\":{\"std_user\":{}}},\"courses_as_student\":{\"elements\":{}},\"courses_as_staff\":{\"elements\":{}}}]}"),
			testUsers[users.Admin],
		},
		{
			"test register users with secretary",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.Users),
			http.StatusAccepted,
			[]byte("{\"users\":[{\"user_name\":\"test2\",\"password\":\"test\",\"roles\":{\"elements\":{\"std_user\":{}}},\"courses_as_student\":{\"elements\":{}},\"courses_as_staff\":{\"elements\":{}}}]}"),
			testUsers[users.Secretary],
		},
		{
			"test register users with std_user",
			http.MethodPost,
			fmt.Sprintf("/%s/", db.Users),
			http.StatusForbidden,
			[]byte("{\"users\":[{\"user_name\":\"test3\",\"password\":\"test\",\"roles\":{\"elements\":{\"std_user\":{}}},\"courses_as_student\":{\"elements\":{}},\"courses_as_staff\":{\"elements\":{}}}]}"),
			testUsers[users.StandardUser],
		},
		{
			"test get user with admin",
			http.MethodGet,
			fmt.Sprintf("/%s/%s", db.Users, users.StandardUser),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
		},
		{
			"test get user with secretary",
			http.MethodGet,
			fmt.Sprintf("/%s/%s", db.Users, users.StandardUser),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Secretary],
		},
		{
			"test get self",
			http.MethodGet,
			fmt.Sprintf("/%s/%s", db.Users, users.StandardUser),
			http.StatusOK,
			[]byte(""),
			testUsers[users.StandardUser],
		},
		{
			"test get admin with std_user",
			http.MethodGet,
			fmt.Sprintf("/%s/%s", db.Users, users.Admin),
			http.StatusForbidden,
			[]byte(""),
			testUsers[users.StandardUser],
		},
		{
			"test update self",
			http.MethodPut,
			fmt.Sprintf("/%s/%s", db.Users, users.StandardUser),
			http.StatusAccepted,
			testUsersBytes[users.StandardUser],
			testUsers[users.StandardUser],
		},
		{
			"test update admin with std_user",
			http.MethodPut,
			fmt.Sprintf("/%s/%s", db.Users, users.Admin),
			http.StatusForbidden,
			testUsersBytes[users.Admin],
			testUsers[users.StandardUser],
		},
		{
			"test update std_user with admin",
			http.MethodPut,
			fmt.Sprintf("/%s/%s", db.Users, users.StandardUser),
			http.StatusAccepted,
			testUsersBytes[users.StandardUser],
			testUsers[users.Admin],
		},
		{
			"test update std_user with secretary",
			http.MethodPut,
			fmt.Sprintf("/%s/%s", db.Users, users.StandardUser),
			http.StatusAccepted,
			testUsersBytes[users.StandardUser],
			testUsers[users.Secretary],
		},
		{
			"test delete admin with std_user",
			http.MethodDelete,
			fmt.Sprintf("/%s/%s", db.Users, users.Admin),
			http.StatusForbidden,
			[]byte(""),
			testUsers[users.StandardUser],
		},
		{
			"test delete std_user with secretary",
			http.MethodDelete,
			fmt.Sprintf("/%s/%s", db.Users, users.StandardUser),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Secretary],
		},
		{
			"test delete secretary with admin",
			http.MethodDelete,
			fmt.Sprintf("/%s/%s", db.Users, users.Secretary),
			http.StatusOK,
			[]byte(""),
			testUsers[users.Admin],
		},
	}
	router := mux.NewRouter()
	am := NewAuthManager()
	router.Use(contentTypeMiddleware, authenticationMiddleware, am.authorizationMiddleware)
	initUsersRouter(router, am)
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
