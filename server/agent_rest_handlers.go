package server

import (
	"encoding/json"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/agents"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"net/http"
)

func handleGetAgents(w http.ResponseWriter, r *http.Request) {
	params, err := submithttp.PagingParamsFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, fmt.Errorf("error parsing query params: %v", err))
		return
	}
	var elementsCount, elementsIndex int64
	var elements []db.IBucketElement
	if err := db.QueryBucket([]byte(db.Agents), func(_, elementBytes []byte) error {
		elementsIndex++
		if elementsIndex <= params.AfterId {
			return nil
		}
		agent := &agents.Agent{}
		if err := json.Unmarshal(elementBytes, agent); err != nil {
			return err
		}
		elements = append(elements, agent)
		elementsCount++
		if elementsCount == params.Limit {
			return &db.ErrStopQuery{}
		}
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
	params, err := submithttp.PagingParamsFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, fmt.Errorf("error parsing query params: %v", err))
		return
	}
	var elementsCount, elementsIndex int64
	var elements []db.IBucketElement
	if err := db.QueryBucket([]byte(db.Tasks), func(_, elementBytes []byte) error {
		elementsIndex++
		if elementsIndex <= params.AfterId {
			return nil
		}
		task := &agents.Task{}
		if err := json.Unmarshal(elementBytes, task); err != nil {
			return err
		}
		elements = append(elements, task)
		elementsCount++
		if elementsCount == params.Limit {
			return &db.ErrStopQuery{}
		}
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
	params, err := submithttp.PagingParamsFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, fmt.Errorf("error parsing query params: %v", err))
		return
	}
	var elementsCount, elementsIndex int64
	var elements []db.IBucketElement
	if err := db.QueryBucket([]byte(db.TaskResponses), func(_, elementBytes []byte) error {
		elementsIndex++
		if elementsIndex <= params.AfterId {
			return nil
		}
		taskResponse := &agents.TaskResponse{}
		if err := json.Unmarshal(elementBytes, taskResponse); err != nil {
			return err
		}
		elements = append(elements, taskResponse)
		elementsCount++
		if elementsCount == params.Limit {
			return &db.ErrStopQuery{}
		}
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

func handlePostTask(w http.ResponseWriter, r *http.Request) {
	task := &agents.Task{}
	if err := json.NewDecoder(r.Body).Decode(task); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	builder := agents.NewTaskBuilder(r.Context().Value(authenticatedUser).(*users.User).UserName, true)
	builder.WithOsType(task.OsType).WithArchitecture(task.Architecture).WithCommand(task.Command).WithResponseHandler(task.ResponseHandler).
		WithExecTimeout(task.ExecTimeout).WithDependencies(task.Dependencies.Slice()...).WithAgent(task.Agent)
	for name, value := range task.Labels {
		builder.WithLabel(name, value)
	}
	task, err := builder.Build()
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{"task created successfully"})
}
