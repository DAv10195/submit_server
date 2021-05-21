package server

import (
	"encoding/json"
	"fmt"
	commons "github.com/DAv10195/submit_commons"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/agents"
	"testing"
	"time"
)

func getDbForMonitorTest() (map[string]string, func()) {
	cleanup := db.InitDbForTest()
	agent1 := &agents.Agent{
		ID:              commons.GenerateUniqueId(),
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
		ID:              commons.GenerateUniqueId(),
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
		ID:              commons.GenerateUniqueId(),
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
		ID:              commons.GenerateUniqueId(),
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
	endpoint1 := newAgentEndpoint(agent1.ID, nil, agent1.User)
	endpoint1.isClosed = true
	agentEndpoints.addEndpoint(endpoint1)
	endpoint2 := newAgentEndpoint(agent2.ID, nil, agent2.User)
	endpoint2.isClosed = true
	agentEndpoints.addEndpoint(endpoint2)
	endpoint3 := newAgentEndpoint(agent3.ID, nil, agent3.User)
	endpoint3.isClosed = true
	agentEndpoints.addEndpoint(endpoint3)
	endpoint4 := newAgentEndpoint(agent4.ID, nil, agent4.User)
	endpoint4.isClosed = true
	agentEndpoints.addEndpoint(endpoint4)
	agentsMap := make(map[string]string)
	agentsMap["windows_amd64"] = agent1.ID
	agentsMap["linux_amd64"] = agent2.ID
	agentsMap["windows_386"] = agent3.ID
	agentsMap["linux_386"] = agent4.ID
	return agentsMap, cleanup
}

func TestStatusMonitor(t *testing.T) {
	agentsMap, cleanup := getDbForMonitorTest()
	defer cleanup()
	agent1, err := agents.Get(agentsMap["linux_amd64"])
	if err != nil {
		t.Fatalf("error getting agent for test: %v", err)
	}
	agent1.LastKeepalive = time.Now().UTC().Add(-3 * time.Minute)
	agent2, err := agents.Get(agentsMap["windows_amd64"])
	if err != nil {
		t.Fatalf("error getting agent for test: %v", err)
	}
	agent2.LastKeepalive = time.Now().UTC().Add(-3 * time.Minute)
	if err := db.Update(db.System, agent1, agent2); err != nil {
		t.Fatalf("error updating agents for test: %v", err)
	}
	agentEndpoints.processAgentsKeepalive()
	if err := db.QueryBucket([]byte(db.Agents), func (_, agentBytes []byte) error {
		agent := &agents.Agent{}
		if err := json.Unmarshal(agentBytes, agent); err != nil {
			return err
		}
		if (agent.ID == agentsMap["linux_amd64"] || agent.ID == agentsMap["windows_amd64"]) && agent.Status != agents.Down {
			t.Fatalf("expected agent with id == %s to be down but he's not", agent.ID)
		}
		if (agent.ID == agentsMap["linux_386"] || agent.ID == agentsMap["windows_386"]) && agent.Status != agents.Up {
			t.Fatalf("expected agent with id == %s to be up but he's not", agent.ID)
		}
		return nil
	}); err != nil {
		t.Fatalf("error querying agents bucket: %v", err)
	}
}

func TestTasksMonitor(t *testing.T) {
	_, cleanup := getDbForMonitorTest()
	defer cleanup()
	// create 25 tasks per agent using filters
	for i := 0; i < 5; i++ {
		builder := agents.NewTaskBuilder(db.System, true)
		builder.WithExecTimeout(1).WithResponseHandler("mock").WithCommand("mock").
			WithArchitecture("amd64").WithOsType("windows")
		if _, err := builder.Build(); err != nil {
			t.Fatalf("error creating task for test: %v", err)
		}
	}
	for i := 0; i < 5; i++ {
		builder := agents.NewTaskBuilder(db.System, true)
		builder.WithExecTimeout(1).WithResponseHandler("mock").WithCommand("mock").
			WithArchitecture("amd64").WithOsType("linux")
		if _, err := builder.Build(); err != nil {
			t.Fatalf("error creating task for test: %v", err)
		}
	}
	for i := 0; i < 5; i++ {
		builder := agents.NewTaskBuilder(db.System, true)
		builder.WithExecTimeout(1).WithResponseHandler("mock").WithCommand("mock").
			WithArchitecture("386").WithOsType("windows")
		if _, err := builder.Build(); err != nil {
			t.Fatalf("error creating task for test: %v", err)
		}
	}
	for i := 0; i < 5; i++ {
		builder := agents.NewTaskBuilder(db.System, true)
		builder.WithExecTimeout(1).WithResponseHandler("mock").WithCommand("mock").
			WithArchitecture("386").WithOsType("linux")
		if _, err := builder.Build(); err != nil {
			t.Fatalf("error creating task for test: %v", err)
		}
	}
	agentEndpoints.processTasks()
	// check that all tasks are in progress
	var tasks []*agents.Task
	if err := db.QueryBucket([]byte(db.Tasks), func (_, taskBytes []byte) error {
		task := &agents.Task{}
		if err := json.Unmarshal(taskBytes, task); err != nil {
			return err
		}
		if task.Status != agents.TaskStatusInProgress {
			t.Fatalf("task for agent with id == %s with id == %s should be in progress but it isn't", task.Agent, task.ID)
		}
		tasks = append(tasks, task)
		return nil
	}); err != nil {
		t.Fatalf("error querying tasks bucket for test: %v", err)
	}
	type mockResp struct {
		Message string `json:"message"`
	}
	agentTaskRespHandlers["mock"] = func (payload []byte) error {
		mr := &mockResp{}
		if err := json.Unmarshal(payload, mr); err != nil {
			return err
		}
		return nil
	}
	// create a response for each task
	var taskElementsToUpdate []db.IBucketElement
	for _, task := range tasks {
		payload, err := json.Marshal(&mockResp{fmt.Sprintf("response for task %s", task.ID)})
		if err != nil {
			t.Fatalf("error preparing task response for test: %v", err)
		}
		tr := &agents.TaskResponse{
			ID:             commons.GenerateUniqueId(),
			Payload:        string(payload),
			Handler:        "mock",
			Task:           task.ID,
		}
		task.TaskResponse = tr.ID
		taskElementsToUpdate = append(taskElementsToUpdate, task, tr)
	}
	if err := db.Update(db.System, taskElementsToUpdate...); err != nil {
		t.Fatalf("error updating task and responses for test: %v", err)
	}
	agentEndpoints.processTasks()
	// check that all tasks are done
	if err := db.QueryBucket([]byte(db.Tasks), func (_, taskBytes []byte) error {
		task := &agents.Task{}
		if err := json.Unmarshal(taskBytes, task); err != nil {
			return err
		}
		if task.Status != agents.TaskStatusDone {
			t.Fatalf("task for agent with id == %s with id == %s should be done but it isn't", task.Agent, task.ID)
		}
		return nil
	}); err != nil {
		t.Fatalf("error querying tasks bucket for test: %v", err)
	}
}
