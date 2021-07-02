package server

import (
	"bytes"
	"encoding/json"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/session"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func getDbForSessionTest() func() {
	sessionCleanup, dbCleanUp := session.InitSessionForTest(), db.InitDbForTest()
	if err := users.InitDefaultAdmin(); err != nil {
		panic(err)
	}
	return func() {
		sessionCleanup()
		dbCleanUp()
	}
}

func TestSession(t *testing.T) {
	cleanup := getDbForSessionTest()
	defer cleanup()
	router := mux.NewRouter()
	am := NewAuthManager()
	router.Use(contentTypeMiddleware, authenticationMiddleware, am.authorizationMiddleware)
	router.HandleFunc("/", session.LoginHandler(nil))
	r, err := http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Fatalf("error creating request for test: %v", err)
	}
	r.SetBasicAuth(users.Admin, users.Admin)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("error calling login handler. Expected 200 status code but got %d", w.Code)
	}
	ld := &session.LoginData{}
	if err := json.NewDecoder(w.Body).Decode(ld); err != nil {
		t.Fatal("error parsing login data from response")
	}
	if ld.UserName != users.Admin {
		t.Fatalf("expected login data user to be admin but got %s", ld.UserName)
	}
	if len(ld.Roles) != 1 || ld.Roles[0] != users.Admin {
		t.Fatalf("expected roles in login data to have a single admin role but got %v", ld.Roles)
	}
	if ld.StaffCourses != nil {
		t.Fatalf("expected staff courses in login data to be nil but got %v", ld.StaffCourses)
	}
	if ld.StudentCourses != nil {
		t.Fatalf("expected student courses in login data to be nil but got %v", ld.StudentCourses)
	}
}
