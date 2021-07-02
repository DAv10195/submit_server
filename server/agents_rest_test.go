package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/DAv10195/submit_commons"
	"github.com/DAv10195/submit_commons/containers"
	submitws "github.com/DAv10195/submit_commons/websocket"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/agents"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/session"
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
	cleanupSess := session.InitSessionForTest()
	defer cleanupSess()
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

func TestTaskRestHandlers(t *testing.T) {
	cleanup := db.InitDbForTest()
	defer cleanup()
	if err := users.InitDefaultAdmin(); err != nil {
		t.Fatalf("error initialiting admin user for test: %v", err)
	}
	taskId := submit_commons.GenerateUniqueId()
	task := &agents.Task{
		ID: taskId,
		Command: "mock",
		ResponseHandler: "mock",
		ExecTimeout: 1,
		Status: agents.TaskStatusDone,
		Dependencies: containers.NewStringSet(),
	}
	if err := db.Update(db.System, task); err != nil {
		t.Fatalf("error updating db with task for test: %v", err)
	}
	task.ID = ""
	task.Status = 0
	taskBytes, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("error seralizing task for test: %v", err)
	}
	if _, err := users.NewUserBuilder(db.System, true).WithUserName(users.Secretary).WithPassword(users.Secretary).WithRoles(users.Secretary).Build(); err != nil {
		t.Fatalf("error creating secretary user for agent rest test: %v", err)
	}
	allTasksPath, taskPath := fmt.Sprintf("/%s/", db.Tasks), fmt.Sprintf("/%s/%s", db.Tasks, taskId)
	testCases := []struct{
		name	string
		method	string
		path	string
		user	string
		data	[]byte
		status	int
	}{
		{
			"test get all tasks",
			http.MethodGet,
			allTasksPath,
			users.Admin,
			[]byte(""),
			http.StatusOK,
		},
		{
			"test get specific task",
			http.MethodGet,
			taskPath,
			users.Admin,
			[]byte(""),
			http.StatusOK,
		},
		{
			"test get all tasks forbidden",
			http.MethodGet,
			allTasksPath,
			users.Secretary,
			[]byte(""),
			http.StatusForbidden,
		},
		{
			"test get specific task forbidden",
			http.MethodGet,
			taskPath,
			users.Secretary,
			[]byte(""),
			http.StatusForbidden,
		},
		{
			"test get all tasks invalid method",
			http.MethodDelete,
			allTasksPath,
			users.Secretary,
			[]byte("bad"),
			http.StatusMethodNotAllowed,
		},
		{
			"test get specific task invalid method",
			http.MethodPost,
			taskPath,
			users.Secretary,
			[]byte("bad"),
			http.StatusMethodNotAllowed,
		},
		{
			"post task",
			http.MethodPost,
			allTasksPath,
			users.Admin,
			taskBytes,
			http.StatusAccepted,
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

func TestTaskRespRestHandlers(t *testing.T) {
	cleanup := db.InitDbForTest()
	defer cleanup()
	if err := users.InitDefaultAdmin(); err != nil {
		t.Fatalf("error initialiting admin user for test: %v", err)
	}
	taskRespId := submit_commons.GenerateUniqueId()
	taskResponse := &agents.TaskResponse{
		ID: taskRespId,
		Payload: "mock",
		Handler: "mock",
	}
	if err := db.Update(db.System, taskResponse); err != nil {
		t.Fatalf("error updating db with agent for test: %v", err)
	}
	if _, err := users.NewUserBuilder(db.System, true).WithUserName(users.Secretary).WithPassword(users.Secretary).WithRoles(users.Secretary).Build(); err != nil {
		t.Fatalf("error creating secretary user for agent rest test: %v", err)
	}
	allTasksPath, taskPath := fmt.Sprintf("/%s/", db.TaskResponses), fmt.Sprintf("/%s/%s", db.TaskResponses, taskRespId)
	testCases := []struct{
		name	string
		method	string
		path	string
		user	string
		data	[]byte
		status	int
	}{
		{
			"test get all tasks responses",
			http.MethodGet,
			allTasksPath,
			users.Admin,
			[]byte(""),
			http.StatusOK,
		},
		{
			"test get specific task response",
			http.MethodGet,
			taskPath,
			users.Admin,
			[]byte(""),
			http.StatusOK,
		},
		{
			"test get all task responses forbidden",
			http.MethodGet,
			allTasksPath,
			users.Secretary,
			[]byte(""),
			http.StatusForbidden,
		},
		{
			"test get specific task response forbidden",
			http.MethodGet,
			taskPath,
			users.Secretary,
			[]byte(""),
			http.StatusForbidden,
		},
		{
			"test get all task responses invalid method",
			http.MethodPost,
			allTasksPath,
			users.Secretary,
			[]byte("bad"),
			http.StatusMethodNotAllowed,
		},
		{
			"test get specific task response invalid method",
			http.MethodPost,
			taskPath,
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
