package server

import (
	"encoding/json"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/agents"
	"github.com/gorilla/mux"
	"net/http"
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

func handleGetTasks(w http.ResponseWriter, r *http.Request) {
	var elements []db.IBucketElement
	if err := db.QueryBucket([]byte(db.Tasks), func(_, elementBytes []byte) error {
		task := &agents.Task{}
		if err := json.Unmarshal(elementBytes, task); err != nil {
			return err
		}
		elements = append(elements, task)
		return nil
	}); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeElements(w, r, http.StatusOK, elements)
}

func handleGetTask(w http.ResponseWriter, r *http.Request) {
	requestedTaskId := mux.Vars(r)[taskId]
	requestedTask, err := agents.GetTask(requestedTaskId)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	writeElem(w, r, http.StatusOK, requestedTask)
}

func handleGetTaskResponses(w http.ResponseWriter, r *http.Request) {
	var elements []db.IBucketElement
	if err := db.QueryBucket([]byte(db.TaskResponses), func(_, elementBytes []byte) error {
		taskResponse := &agents.TaskResponse{}
		if err := json.Unmarshal(elementBytes, taskResponse); err != nil {
			return err
		}
		elements = append(elements, taskResponse)
		return nil
	}); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeElements(w, r, http.StatusOK, elements)
}

func handleGetTaskResponse(w http.ResponseWriter, r *http.Request) {
	requestedTaskResponseId := mux.Vars(r)[taskRespId]
	requestedTaskResponse, err := agents.GetTaskResponse(requestedTaskResponseId)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	writeElem(w, r, http.StatusOK, requestedTaskResponse)
}
