package server

import (
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
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
