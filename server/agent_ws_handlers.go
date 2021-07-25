package server

import (
	"encoding/json"
	commons "github.com/DAv10195/submit_commons"
	submitws "github.com/DAv10195/submit_commons/websocket"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/agents"
	"time"
)

type agentMessageHandler func(string, []byte)

var agentMsgHandlers = make(map[string]agentMessageHandler)

// handle keepalive messages from agents
func handleKeepalive(agentId string, payload []byte) {
	logger.Debugf("keepalive handler: received keepalive message [ %s ] from agent with id == %s", string(payload), agentId)
	var endpoint *agentEndpoint
	if endpoint = agentEndpoints.getEndpoint(agentId); endpoint != nil {
		keepaliveResp := &submitws.KeepaliveResponse{Message: hello}
		keepaliveRespBytes, err := json.Marshal(keepaliveResp)
		if err != nil {
			logger.WithError(err).Errorf("keepalive handler: error formatting keepalive response")
		}
		msg, err := submitws.NewMessage(submitws.MessageTypeKeepaliveResponse, keepaliveRespBytes)
		if err != nil {
			logger.WithError(err).Errorf("keepalive handler: error creating keepalive response message")
		}
		endpoint.write(msg)
	} else {
		logger.Warnf("keepalive handler: no endpoint for agent with id == %s", agentId)
		return
	}
	keepalive := &submitws.Keepalive{}
	err := json.Unmarshal(payload, keepalive)
	if err != nil {
		logger.WithError(err).Error("keepalive handler: error parsing keepalive message")
		return
	}
	var agent *agents.Agent
	if agent, err = agents.Get(agentId); err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); !ok {
			logger.WithError(err).Errorf("keepalive handler: error querying for agent with id == %s", agentId)
			return
		}
		agent = &agents.Agent{
			ID: agentId,
		}
	}
	agent.User = endpoint.user
	agent.Hostname = keepalive.Hostname
	agent.IpAddress = keepalive.IpAddress
	agent.OsType = keepalive.OsType
	agent.Architecture = keepalive.Architecture
	agent.NumRunningTasks = keepalive.NumRunningTasks
	agent.Status = agents.Up
	agent.LastKeepalive = time.Now().UTC()
	if err = db.Update(endpoint.user, agent); err != nil {
		logger.WithError(err).Errorf("keepalive handler: error updating agent with id == %s in the db", agentId)
	}
}

// handle task responses from agents - update the task with the response and move it to done status so the
// processing job will pick it up and process it
func handleTaskResponses(agentId string, payload []byte) {
	logger.Debugf("task responses handler: received task response message [ %s ] from agent with id == %s", string(payload), agentId)
	var endpoint *agentEndpoint
	if endpoint = agentEndpoints.getEndpoint(agentId); endpoint == nil {
		logger.Warnf("task responses handler: no endpoint for agent with id == %s", agentId)
		return
	}
	taskResponsesFromAgent := &submitws.TaskResponses{}
	if err := json.Unmarshal(payload, taskResponsesFromAgent); err != nil {
		logger.WithError(err).Error("task responses handler: error parsing task response message")
		return
	}
	for _, taskResponseFromAgent := range taskResponsesFromAgent.Responses {
		task, err := agents.GetTask(taskResponseFromAgent.Task)
		if err != nil {
			logger.WithError(err).Errorf("task responses handler: received response for task with id == %s but it doesn't exist", taskResponseFromAgent.Task)
			continue
		}
		if task.Status != agents.TaskStatusInProgress {
			logger.Warnf("ignoring response for task with id == '%s' as it is not in progress: %s", task.ID, string(payload))
			continue
		}
		taskResponse := &agents.TaskResponse{
			ID:             commons.GenerateUniqueId(),
			Payload:        taskResponseFromAgent.Payload,
			Handler:        taskResponseFromAgent.Handler,
			Task:           taskResponseFromAgent.Task,
			ExecStatus:		taskResponseFromAgent.Status,
			Labels:			taskResponseFromAgent.Labels,
		}
		task.TaskResponse = taskResponse.ID
		task.Status = agents.TaskStatusDone
		if err := db.Update(endpoint.user, taskResponse, task); err != nil {
			logger.WithError(err).Error("task responses handler: error updating task and response")
		}
	}
}

func init() {
	agentMsgHandlers[submitws.MessageTypeKeepalive] = handleKeepalive
	agentMsgHandlers[submitws.MessageTypeTaskResponses] = handleTaskResponses
}
