package server

import (
	"context"
	"encoding/json"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/agents"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/session"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func getDbForPaginationTest() func() {
	cleanup := db.InitDbForTest()
	agent1 := &agents.Agent{
		ID:              "agent1",
		User:            "agent1",
		Hostname:        "agent1",
		IpAddress:       "1.1.1.1",
		OsType:          "windows",
		Architecture:    "amd64",
		Status:          agents.Up,
		NumRunningTasks: 0,
		LastKeepalive:   time.Now().UTC(),
	}
	agent2 := &agents.Agent{
		ID:              "agent2",
		User:            "agent2",
		Hostname:        "agent2",
		IpAddress:       "2.2.2.2",
		OsType:          "linux",
		Architecture:    "amd64",
		Status:          agents.Up,
		NumRunningTasks: 0,
		LastKeepalive:   time.Now().UTC(),
	}
	agent3 := &agents.Agent{
		ID:              "agent3",
		User:            "agent3",
		Hostname:        "agent3",
		IpAddress:       "3.3.3.3",
		OsType:          "windows",
		Architecture:    "386",
		Status:          agents.Up,
		NumRunningTasks: 0,
		LastKeepalive:   time.Now().UTC(),
	}
	agent4 := &agents.Agent{
		ID:              "agent4",
		User:            "agent4",
		Hostname:        "agent4",
		IpAddress:       "4.4.4.4",
		OsType:          "linux",
		Architecture:    "386",
		Status:          agents.Up,
		NumRunningTasks: 0,
		LastKeepalive:   time.Now().UTC(),
	}
	if err := db.Update(db.System, agent1, agent2, agent3, agent4); err != nil {
		cleanup()
		panic(err)
	}
	return cleanup
}

func TestRestPagination(t *testing.T) {
	cleanup := getDbForPaginationTest()
	defer cleanup()
	cleanupSess := session.InitSessionForTest()
	defer cleanupSess()
	if err := users.InitDefaultAdmin(); err != nil {
		t.Fatalf("error initialiting admin user for test: %v", err)
	}
	router := mux.NewRouter()
	am := NewAuthManager()
	router.Use(contentTypeMiddleware, authenticationMiddleware, am.authorizationMiddleware)
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initAgentsBackend(router, am, ctx, wg)
	r, err := http.NewRequest(http.MethodGet, "/agents/?limit=1", nil)
	if err != nil {
		t.Fatalf("error creating http request for test: %v", err)
	}
	r.SetBasicAuth(users.Admin, users.Admin)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d but got %d", http.StatusOK, w.Code)
	}
	type body struct {
		Agents []*agents.Agent `json:"elements"`
	}
	respBody := &body{}
	if err := json.NewDecoder(w.Body).Decode(respBody); err != nil {
		t.Fatalf("error parsing response body for test: %v", err)
	}
	if len(respBody.Agents) != 1 {
		t.Fatalf("expected 1 element in reponse but got %d", len(respBody.Agents))
	}
	agentId := respBody.Agents[0].ID
	if agentId != "agent1" {
		t.Fatalf("expected agent1 in reponse but got %s", agentId)
	}
	if w.Header().Get(submithttp.ElementsLeftToProcess) != trueStr {
		t.Fatalf("expected header %s to be %s but it is %s", submithttp.ElementsLeftToProcess, trueStr, w.Header().Get(submithttp.ElementsLeftToProcess))
	}
	r, err = http.NewRequest(http.MethodGet, "/agents/?limit=1&after_id=1", nil)
	if err != nil {
		t.Fatalf("error creating http request for test: %v", err)
	}
	r.SetBasicAuth(users.Admin, users.Admin)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d but got %d", http.StatusOK, w.Code)
	}
	respBody = &body{}
	if err := json.NewDecoder(w.Body).Decode(respBody); err != nil {
		t.Fatalf("error parsing response body for test: %v", err)
	}
	if len(respBody.Agents) != 1 {
		t.Fatalf("expected 1 element in reponse but got %d", len(respBody.Agents))
	}
	agentId = respBody.Agents[0].ID
	if agentId != "agent2" {
		t.Fatalf("expected agent1 in reponse but got %s", agentId)
	}
	if w.Header().Get(submithttp.ElementsLeftToProcess) != trueStr {
		t.Fatalf("expected header %s to be %s but it is %s", submithttp.ElementsLeftToProcess, trueStr, w.Header().Get(submithttp.ElementsLeftToProcess))
	}
	r, err = http.NewRequest(http.MethodGet, "/agents/?limit=1&after_id=2", nil)
	if err != nil {
		t.Fatalf("error creating http request for test: %v", err)
	}
	r.SetBasicAuth(users.Admin, users.Admin)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d but got %d", http.StatusOK, w.Code)
	}
	respBody = &body{}
	if err := json.NewDecoder(w.Body).Decode(respBody); err != nil {
		t.Fatalf("error parsing response body for test: %v", err)
	}
	if len(respBody.Agents) != 1 {
		t.Fatalf("expected 1 element in reponse but got %d", len(respBody.Agents))
	}
	agentId = respBody.Agents[0].ID
	if agentId != "agent3" {
		t.Fatalf("expected agent1 in reponse but got %s", agentId)
	}
	if w.Header().Get(submithttp.ElementsLeftToProcess) != trueStr {
		t.Fatalf("expected header %s to be %s but it is %s", submithttp.ElementsLeftToProcess, trueStr, w.Header().Get(submithttp.ElementsLeftToProcess))
	}
	r, err = http.NewRequest(http.MethodGet, "/agents/?limit=1&after_id=3", nil)
	if err != nil {
		t.Fatalf("error creating http request for test: %v", err)
	}
	r.SetBasicAuth(users.Admin, users.Admin)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d but got %d", http.StatusOK, w.Code)
	}
	respBody = &body{}
	if err := json.NewDecoder(w.Body).Decode(respBody); err != nil {
		t.Fatalf("error parsing response body for test: %v", err)
	}
	if len(respBody.Agents) != 1 {
		t.Fatalf("expected 1 element in reponse but got %d", len(respBody.Agents))
	}
	agentId = respBody.Agents[0].ID
	if agentId != "agent4" {
		t.Fatalf("expected agent1 in reponse but got %s", agentId)
	}
	if w.Header().Get(submithttp.ElementsLeftToProcess) != "" {
		t.Fatalf("expected header %s to be empty but it is %s", submithttp.ElementsLeftToProcess, w.Header().Get(submithttp.ElementsLeftToProcess))
	}
}
