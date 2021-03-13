package server

import (
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/messages"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/util/containers"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func getRouterForMiddlewareTest() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("\"message\": \"hello from fake router\"")); err != nil {
			panic(err)
		}
	})
	return router
}

func TestContentTypeMiddleware(t *testing.T) {
	router := getRouterForMiddlewareTest()
	router.Use(contentTypeMiddleware)
	request, err := http.NewRequest("", "/", nil)
	if err != nil {
		t.Fatalf("error creating http request: %v", err)
	}
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)
	contentTypeHeaderValues := writer.Result().Header[ContentType]
	if len(contentTypeHeaderValues) != 1 && contentTypeHeaderValues[0] != ApplicationJson {
		t.Fatalf("content type header has more than 1 element or it doesn't match %s: %v", ApplicationJson, contentTypeHeaderValues)
	}
}

func TestAuthenticationMiddleware(t *testing.T) {
	cleanup := db.InitDbForTest()
	defer cleanup()
	if err := users.InitDefaultAdmin(); err != nil {
		t.Fatalf("error initializng default admin user: %v", err)
	}
	router := getRouterForMiddlewareTest()
	router.Use(authenticationMiddleware)
	request, err := http.NewRequest("", "/", nil)
	if err != nil {
		t.Fatalf("error creating http request: %v", err)
	}
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)
	if writer.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d but got %d", http.StatusUnauthorized, writer.Code)
	}
	request, err = http.NewRequest("", "/", nil)
	if err != nil {
		t.Fatalf("error creating http request: %v", err)
	}
	request.SetBasicAuth(users.Admin, "password")
	writer = httptest.NewRecorder()
	router.ServeHTTP(writer, request)
	if writer.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d but got %d", http.StatusUnauthorized, writer.Code)
	}
	request, err = http.NewRequest("", "/", nil)
	if err != nil {
		t.Fatalf("error creating http request: %v", err)
	}
	request.SetBasicAuth(users.Admin, users.Admin)
	writer = httptest.NewRecorder()
	router.ServeHTTP(writer, request)
	if writer.Code != http.StatusOK {
		t.Fatalf("expected status %d but got %d", http.StatusOK, writer.Code)
	}
}

func TestAuthorizationMiddleware(t *testing.T){
	cleanup := db.InitDbForTest()
	defer cleanup()
	am := InitAuthManager()
	initTestAuthManager(am)
	if err := users.InitDefaultAdmin(); err != nil {
		t.Fatalf("error initializng default admin user: %v", err)
	}
	router := getRouterForMiddlewareTest()
	router.Use(authenticationMiddleware)
	router.Use(am.authorizationMiddleware)
	router.HandleFunc("/regex/{suffix}", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("\"message\": \"hello from /regex/")); err != nil {
			panic(err)
		}
	})
	// use admin user for positive test.
	request, err := http.NewRequest("", "/", nil)
	if err != nil {
		t.Fatalf("error creating http request: %v", err)
	}
	request.SetBasicAuth(users.Admin, users.Admin)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)
	if writer.Code != http.StatusOK {
		t.Fatalf("expected status %d but got %d", http.StatusOK, writer.Code)
	}
	request, err = http.NewRequest("", "/regex/test", nil)
	if err != nil {
		t.Fatalf("error creating http request: %v", err)
	}
	request.SetBasicAuth(users.Admin, users.Admin)
	writer = httptest.NewRecorder()
	router.ServeHTTP(writer, request)
	if writer.Code != http.StatusOK {
		t.Fatalf("expected status %d but got %d", http.StatusOK, writer.Code)
	}
	// register a new user and try to access the content protected by auth manager.
	messageBox := messages.NewMessageBox()
	encryptedPassword, err := db.Encrypt("nikita")
	if err != nil {
		t.Fatalf("failed to encrypr the password")
	}
	userNikita := &users.User{
		UserName: "nikita",
		Password: encryptedPassword,
		MessageBox: "messageBox",
		Roles: containers.NewStringSet(),
		CoursesAsStaff: containers.NewStringSet(),
		CoursesAsStudent: containers.NewStringSet(),
	}
	userNikita.Roles.Add(users.Admin)
	err = db.Update(db.System, messageBox, userNikita)
	if err != nil {
		t.Fatalf("error updating db : %v", err)
	}
	request, err = http.NewRequest("", "/", nil)
	if err != nil {
		t.Fatalf("error creating http request: %v", err)
	}
	request.SetBasicAuth("nikita", "nikita")
	writer = httptest.NewRecorder()
	router.ServeHTTP(writer, request)
	if writer.Code != http.StatusForbidden {
		t.Fatalf("expected status %d but got %d", http.StatusForbidden, writer.Code)
	}
	request, err = http.NewRequest("", "/regex/test", nil)
	if err != nil {
		t.Fatalf("error creating http request: %v", err)
	}
	request.SetBasicAuth("nikita", "nikita")
	writer = httptest.NewRecorder()
	router.ServeHTTP(writer, request)
	if writer.Code != http.StatusForbidden {
		t.Fatalf("expected status %d but got %d", http.StatusForbidden, writer.Code)
	}
}

func initTestAuthManager(authManager *authManager){
	authManager.addPathToMap("/", func(user *users.User) bool{
		if user.UserName == "admin" {
			return true
		}
		return false
	})
	regex, _ := regexp.Compile("/regex/.")
	authManager.addRegex(regex, func(user *users.User) bool{
		if user.UserName == "admin" {
			return true
		}
		return false
	})
}
