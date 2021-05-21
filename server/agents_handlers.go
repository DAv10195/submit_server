package server

import (
	"encoding/json"
	commons "github.com/DAv10195/submit_commons"
	submitws "github.com/DAv10195/submit_commons/websocket"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/agents"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func handleGetAgents(w http.ResponseWriter, r *http.Request) {
	var elements []db.IBucketElement
	if err := db.QueryBucket([]byte(db.Agents), func(_, elementBytes []byte) error {
		agent := &agents.Agent{}
		if err := json.Unmarshal(elementBytes, agent); err != nil {
			return err
		}
		elements = append(elements, agent)
		return nil
	}); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeElements(w, r, http.StatusOK, elements)
}

func handleGetAgent(w http.ResponseWriter, r *http.Request) {
	requestedAgentId := mux.Vars(r)[agentId]
	requestedAgent, err := agents.Get(requestedAgentId)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	writeElem(w, r, http.StatusOK, requestedAgent)
}

type agentMessageHandler func(string, []byte)

var agentMsgHandlers = make(map[string]agentMessageHandler)

type agentTaskResponseHandler func([]byte) error

var agentTaskRespHandlers = make(map[string]agentTaskResponseHandler)

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

func handleTaskResponse(agentId string, payload []byte) {
	logger.Debugf("task response handler: received task response message [ %s ] from agent with id == %s", string(payload), agentId)
	var endpoint *agentEndpoint
	if endpoint = agentEndpoints.getEndpoint(agentId); endpoint == nil {
		logger.Warnf("task response handler: no endpoint for agent with id == %s", agentId)
		return
	}
	taskResponseMsg := &submitws.TaskResponse{}
	if err := json.Unmarshal(payload, taskResponseMsg); err != nil {
		logger.WithError(err).Error("task response handler: error parsing task response message")
		return
	}
	task, err := agents.GetTask(taskResponseMsg.Task)
	if err != nil {
		logger.WithError(err).Errorf("task response handler: received response for task with id == %s but it doesn't exist", taskResponseMsg.Task)
		return
	}
	taskResponse := &agents.TaskResponse{
		ID:             commons.GenerateUniqueId(),
		Payload:        taskResponseMsg.Payload,
		Handler:        taskResponseMsg.Handler,
		Task:           taskResponseMsg.Task,
	}
	task.TaskResponse = taskResponse.ID
	if err := db.Update(endpoint.user, taskResponse, task); err != nil {
		logger.WithError(err).Error("task response handler: error updating task and response")
	}
}

func init() {
	agentMsgHandlers[submitws.MessageTypeKeepalive] = handleKeepalive
	agentMsgHandlers[submitws.MessageTypeTaskResponse] = handleTaskResponse
}
