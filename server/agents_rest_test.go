package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/DAv10195/submit_commons"
	submitws "github.com/DAv10195/submit_commons/websocket"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/agents"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestAgentRestHandlers(t *testing.T) {
	cleanup := db.InitDbForTest()
	defer cleanup()
	if err := users.InitDefaultAdmin(); err != nil {
		t.Fatalf("error initialiting admin user for test: %v", err)
	}
	agentId := submit_commons.GenerateUniqueId()
	agent := &agents.Agent{
		ID: agentId,
		User: users.Admin,
		Hostname: "host",
		IpAddress: "0.0.0.0",
		OsType: runtime.GOOS,
		Architecture: runtime.GOARCH,
		NumRunningTasks: 0,
		Status: agents.Down,
		LastKeepalive: time.Now().UTC(),
	}
	if err := db.Update(db.System, agent); err != nil {
		t.Fatalf("error updating db with agent for test: %v", err)
	}
	if _, err := users.NewUserBuilder(db.System, true).WithUserName(users.Secretary).WithPassword(users.Secretary).WithRoles(users.Secretary).Build(); err != nil {
		t.Fatalf("error creating secretary user for agent rest test: %v", err)
	}
	allAgentsPath, agentPath := fmt.Sprintf("/%s/", submitws.Agents), fmt.Sprintf("/%s/%s", submitws.Agents, agentId)
	testCases := []struct{
		name	string
		method	string
		path	string
		user	string
		data	[]byte
		status	int
	}{
		{
			"test get all agents",
			http.MethodGet,
			allAgentsPath,
			users.Admin,
			[]byte(""),
			http.StatusOK,
		},
		{
			"test get specific agent",
			http.MethodGet,
			agentPath,
			users.Admin,
			[]byte(""),
			http.StatusOK,
		},
		{
			"test get all agents forbidden",
			http.MethodGet,
			allAgentsPath,
			users.Secretary,
			[]byte(""),
			http.StatusForbidden,
		},
		{
			"test get specific agent forbidden",
			http.MethodGet,
			agentPath,
			users.Secretary,
			[]byte(""),
			http.StatusForbidden,
		},
		{
			"test get all agents invalid method",
			http.MethodPost,
			allAgentsPath,
			users.Secretary,
			[]byte("bad"),
			http.StatusMethodNotAllowed,
		},
		{
			"test get specific agent invalid method",
			http.MethodPost,
			agentPath,
			users.Secretary,
			[]byte("bad"),
			http.StatusMethodNotAllowed,
		},

	}
	router := mux.NewRouter()
	am := NewAuthManager()
	router.Use(contentTypeMiddleware, authenticationMiddleware, am.authorizationMiddleware)
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initAgentsBackend(router, am, ctx, wg)
	for _, testCase := range testCases {
		var testCaseErr error
		if !t.Run(testCase.name, func (t *testing.T) {
			r, err := http.NewRequest(testCase.method, testCase.path, bytes.NewBuffer(testCase.data))
			if err != nil {
				testCaseErr = fmt.Errorf("error creating http request for test case [ %s ]: %v", testCase.name, err)
				t.FailNow()
			}
			r.SetBasicAuth(testCase.user, testCase.user)
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