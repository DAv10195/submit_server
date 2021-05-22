package server

import (
	"context"
	"encoding/json"
	"fmt"
	commons "github.com/DAv10195/submit_commons"
	submitws "github.com/DAv10195/submit_commons/websocket"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/agents"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"math"
	"net/http"
	"regexp"
	"sort"
	"sync"
	"time"
)

var agentEndpoints *agentEndpointsManager

// a websocket endpoint representing a connection to an agent
type agentEndpoint struct {
	id       string
	conn     *websocket.Conn
	mutex    *sync.RWMutex
	user	 string
	isClosed bool
}

// create a new agent endpoint
func newAgentEndpoint(id string, conn *websocket.Conn, user string) *agentEndpoint {
	return &agentEndpoint{id, conn, &sync.RWMutex{}, user, false}
}

// read incoming messages from an agent. This function should be called only by a single goroutine
func (e *agentEndpoint) readLoop() {
	for {
		wsMsgType, payload, err := e.conn.ReadMessage()
		if err != nil {
			e.mutex.Lock()
			if !e.isClosed {
				logger.WithError(err).Errorf("error reading websocket message from agent with id == %s", e.id)
				if err := e.conn.Close(); err != nil {
					logger.WithError(err).Errorf("error closing connection to agent with id == %s after read error: %v", e.id, err)
				}
				e.isClosed = true
			}
			e.mutex.Unlock()
			return
		}
		if wsMsgType != websocket.BinaryMessage {
			logger.Warnf("invalid message sent from agent with id == %s. websocket message is not a binary message (%d)", e.id, websocket.BinaryMessage)
			continue
		}
		msg, err := submitws.FromBinary(payload)
		if err != nil {
			logger.WithError(err).Warnf("invalid message sent from agent with id == %s. Error parsing websocket message: %v", e.id, err)
			continue
		}
		if agentMsgHandlers[msg.Type] == nil {
			logger.WithError(err).Warnf("invalid message sent form agent with id == %s. No handler for message with type == %s", e.id, msg.Type)
			continue
		}
		go agentMsgHandlers[msg.Type](e.id, msg.Payload)
	}
}

// send a message to an agent
func (e *agentEndpoint) write(msg *submitws.Message) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if e.isClosed {
		return
	}
	if err := e.conn.WriteMessage(websocket.BinaryMessage, msg.ToBinary()); err != nil {
		logger.WithError(err).Errorf("error sending message to agent with id == %s: %v", e.id, err)
		if err := e.conn.Close(); err != nil {
			logger.WithError(err).Errorf("error closing connection to agent with id == %s after write error: %v", e.id, err)
		}
		e.isClosed = true
	}
}

// close the connection with the agent by sending a close message
func (e *agentEndpoint) close() {
	e.mutex.Lock()
	defer func() {
		_ = recover()
		e.mutex.Unlock()
	}()
	if e.isClosed {
		return
	}
	if err := e.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "bye bye")); err != nil {
		logger.WithError(err).Errorf("error sending closing message to agent with id == %s", e.id)
	}
	e.isClosed = true
}

// agent endpoints manager
type agentEndpointsManager struct {
	endpoints 	map[string]*agentEndpoint
	mutex		*sync.RWMutex
}

// create an agent endpoints manager
func newAgentEndpointsManager() *agentEndpointsManager {
	return &agentEndpointsManager{make(map[string]*agentEndpoint), &sync.RWMutex{}}
}

// add an endpoint
func (m *agentEndpointsManager) addEndpoint(endpoint *agentEndpoint) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.endpoints[endpoint.id] = endpoint
}

// get the endpoint which is connected to the agent with the given id
func (m *agentEndpointsManager) getEndpoint(agentId string) *agentEndpoint {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.endpoints[agentId]
}

// accept incoming agent connections
func (m *agentEndpointsManager) agentsEndpoint(w http.ResponseWriter, r *http.Request) {
	agentId := r.Header.Get(submitws.AgentIdHeader)
	if len(agentId) != commons.UniqueIdLen {
		writeStrErrResp(w, r, http.StatusBadRequest, fmt.Sprintf("invalid agent ID sent to agent endpoint [ %s ]", agentId))
		return
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	endpoint := m.endpoints[agentId]
	if endpoint != nil {
		endpoint.mutex.RLock()
		isClosed := endpoint.isClosed
		endpoint.mutex.RUnlock()
		if !isClosed {
			writeStrErrResp(w, r, http.StatusBadRequest, fmt.Sprintf("agent with id == %s already exists", agentId))
			return
		}
	}
	wsUpgrade := websocket.Upgrader{}
	conn, err := wsUpgrade.Upgrade(w, r, nil)
	if err != nil {
		logger.WithError(err).Errorf("error upgrading connection from [ %s ] to websocket", r.RemoteAddr)
		return
	}
	logger.Debugf("successfully upgraded connection from [ %s ] to websocket", r.RemoteAddr)
	endpoint = newAgentEndpoint(agentId, conn, r.Context().Value(authenticatedUser).(*users.User).UserName)
	m.endpoints[agentId] = endpoint
	go endpoint.readLoop()
}

// close all agent endpoints
func (m *agentEndpointsManager) close() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	logger.Info("closing all agent endpoints...")
	for _, endpoint := range m.endpoints {
		logger.Infof("closing agent (id == %s) endpoint", endpoint.id)
		endpoint.close()
	}
}

// mark all agent that send keepalive in the last minute as down and close their connections (if present)
func (m *agentEndpointsManager) processAgentsKeepalive() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	logger.Debug("agents status monitor: processing agent keepalives...")
	now := time.Now().UTC()
	var agentsToMarkAsDown []db.IBucketElement
	if err := db.QueryBucket([]byte(db.Agents), func (_, agentBytes []byte) error {
		agent := &agents.Agent{}
		if err := json.Unmarshal(agentBytes, agent); err != nil {
			return err
		}
		if agent.Status == agents.Up && now.Sub(agent.LastKeepalive) > time.Minute {
			agentsToMarkAsDown = append(agentsToMarkAsDown, agent)
		}
		return nil
	}); err != nil {
		logger.WithError(err).Error("agents status monitor: error querying agents bucket for keepalive processing")
	}
	if len(agentsToMarkAsDown) > 0 {
		for _, agentElem := range agentsToMarkAsDown {
			agent := agentElem.(*agents.Agent)
			agent.Status = agents.Down
			if endpoint := m.endpoints[agent.ID]; endpoint != nil {
				endpoint.close()
				delete(m.endpoints, agent.ID)
			}
		}
		if err := db.Update(db.System, agentsToMarkAsDown...); err != nil {
			logger.WithError(err).Error("agents status monitor: error updating agents bucket after keepalive processing")
		}
	}
	logger.Info("agents status monitor: finished processing agent keepalives")
}

// process agents keepalive each minute. Any agent that didn't send a keepalive in the last minute will be marked
// as down and his connection will be terminated (if present)
func (m *agentEndpointsManager) agentStatusMonitor(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	m.processAgentsKeepalive()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
			case <- ticker.C:
				m.processAgentsKeepalive()
			case <- ctx.Done():
				logger.Info("stopping agent status monitor")
				return
		}
	}
}

// given a task, return the ID of the least busy agent that can run it
func (m *agentEndpointsManager) selectAgentForTask(task *agents.Task) (string, error) {
	var relevantAgents []*agents.Agent
	if err := db.QueryBucket([]byte(db.Agents), func (_, agentBytes []byte) error {
		agent := &agents.Agent{}
		if err := json.Unmarshal(agentBytes, agent); err != nil {
			return err
		}
		if agent.Status != agents.Up {
			return nil
		}
		if task.Architecture != "" && task.Architecture != agent.Architecture {
			return nil
		}
		if task.OsType != "" && task.OsType != agent.OsType {
			return nil
		}
		relevantAgents = append(relevantAgents, agent)
		return nil
	}); err != nil {
		return "", err
	}
	if len(relevantAgents) == 0 {
		return "", fmt.Errorf("no connected agents that can run the task")
	}
	sort.Slice(relevantAgents, func (i, j int) bool {
		return relevantAgents[i].NumRunningTasks < relevantAgents[j].NumRunningTasks
	})
	selectedAgent := relevantAgents[0]
	selectedAgent.NumRunningTasks++
	if err := db.Update(db.System, selectedAgent); err != nil {
		return "", err
	}
	return selectedAgent.ID, nil
}

func (m *agentEndpointsManager) updateTaskWithDescriptionToErr(task *agents.Task, description string) {
	task.Status = agents.TaskStatusError
	task.Description = description
	if err := db.Update(db.System, task); err != nil {
		logger.WithError(err).Errorf("failed updating task with id == %s to error status", task.ID)
	}
}

func (m *agentEndpointsManager) updateTaskWithDescriptionToTimeout(task *agents.Task) {
	task.Status = agents.TaskStatusTimeout
	task.Description = "timeout waiting for task response"
	if err := db.Update(db.System, task); err != nil {
		logger.WithError(err).Errorf("failed updating task with id == %s to timeout status", task.ID)
	}
}

func (m *agentEndpointsManager) processTaskWithResponse(task *agents.Task) {
	task.Status = agents.TaskStatusProcessing
	if err := db.Update(db.System, task); err != nil {
		logger.WithError(err).Errorf("error updating task with id = %s to processing status", task.ID)
		return
	}
	resp, err := agents.GetTaskResponse(task.TaskResponse)
	if err != nil {
		logger.WithError(err).Errorf("error getting task response with id == %s for task with id == %s", task.TaskResponse, task.ID)
		m.updateTaskWithDescriptionToErr(task, err.Error())
		return
	}
	handler := agentTaskRespHandlers[resp.Handler]
	if handler == nil {
		logger.Errorf("handler ('%s') of response for task with id == %s not found", resp.Handler, task.ID)
		m.updateTaskWithDescriptionToErr(task, "response handler not found")
		return
	}
	if err := handler([]byte(resp.Payload)); err != nil {
		logger.WithError(err).Errorf("error handling response for task with id == %s", task.ID)
		m.updateTaskWithDescriptionToErr(task, err.Error())
		return
	}
	task.Description = fmt.Sprintf("successfully processed response using the following handler: %s", resp.Handler)
	task.Status = agents.TaskStatusDone
	if err := db.Update(db.System, task); err != nil {
		logger.WithError(err).Errorf("error updating task with id = %s to done status", task.ID)
	}
}

// process a batch of tasks. Executed by a task processing worker goroutine
func (m *agentEndpointsManager) _processTasks(workerNum int, tasks []*agents.Task, wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Debugf("agents tasks monitor: task processing worker #%d processing %d tasks", workerNum, len(tasks))
	now := time.Now().UTC()
	for _, task := range tasks {
		switch task.Status {
			case agents.TaskStatusReady:
				if task.Agent == "" {
					selectedAgentId, err := m.selectAgentForTask(task)
					if err != nil {
						logger.WithError(err).Errorf("agents tasks monitor: failed selecting agent for task with id == %s", task.ID)
						m.updateTaskWithDescriptionToErr(task, err.Error())
						continue
					}
					task.Agent = selectedAgentId
				}
				selectedAgentEndpoint := m.getEndpoint(task.Agent)
				if selectedAgentEndpoint == nil {
					logger.Errorf("agents tasks monitor: agent with id == %s was selected to run task with id == %s but no endpoint available for him", task.Agent, task.ID)
					m.updateTaskWithDescriptionToErr(task, "agent unavailable")
					continue
				}
				msg, err := task.GetWsMessage()
				if err != nil {
					logger.WithError(err).Errorf("error creating message from task with id == %s", task.ID)
					m.updateTaskWithDescriptionToErr(task, err.Error())
					continue
				}
				task.Status = agents.TaskStatusInProgress
				if err := db.Update(db.System, task); err != nil {
					logger.WithError(err).Errorf("error updating task with id == %s to in progress state", task.ID)
					m.updateTaskWithDescriptionToErr(task, err.Error())
					continue
				}
				selectedAgentEndpoint.write(msg)
			case agents.TaskStatusInProgress:
				if task.TaskResponse != "" {
					m.processTaskWithResponse(task)
				} else if now.Sub(task.UpdatedOn) > taskProcessingTimeout {
					m.updateTaskWithDescriptionToTimeout(task)
				}
			default: // should not happen...
				logger.Errorf("task with id == %s was selected for processing it has status = %d (not in progress or ready)", task.ID, task.Status)
		}
	}
}

// process tasks using processing workers (goroutines)
func (m *agentEndpointsManager) processTasks(wg *sync.WaitGroup) {
	var tasks []*agents.Task
	var taskElementsToDel []db.IBucketElement
	now := time.Now().UTC()
	if err := db.QueryBucket([]byte(db.Tasks), func (_, taskBytes []byte) error {
		task := &agents.Task{}
		if err := json.Unmarshal(taskBytes, task); err != nil {
			return err
		}
		if task.Status == agents.TaskStatusReady || task.Status == agents.TaskStatusInProgress {
			tasks = append(tasks, task)
		} else if now.Sub(task.UpdatedOn) > 7 * 24 * time.Hour { // delete if updated more than a week ago
			taskElementsToDel = append(taskElementsToDel, task)
		}
		return nil
	}); err != nil {
		logger.WithError(err).Error("agents tasks monitor: error querying for tasks to process")
		return
	}
	// if any tasks to delete, then do it in a separate goroutine to not halt the processing and also delete responses...
	if len(taskElementsToDel) > 0 {
		go func() {
			var taskResponseIdsToDel [][]byte
			for _, taskElem := range taskElementsToDel {
				taskResponseId := taskElem.(*agents.Task).TaskResponse
				if taskResponseId != "" {
					taskResponseIdsToDel = append(taskResponseIdsToDel, []byte(taskResponseId))
				}
			}
			if len(taskResponseIdsToDel) > 0 {
				if err := db.DeleteKeysFromBucket([]byte(db.TaskResponses), taskResponseIdsToDel...); err != nil {
					logger.WithError(err).Error("error deleting old task responses (updated more then a week ago)")
				}
			}
			if err := db.Delete(taskElementsToDel...); err != nil {
				logger.WithError(err).Error("error deleting old tasks (updated more then a week ago)")
			}
		}()
	}
	// divide tasks between workers
	numTasks := len(tasks)
	if numTasks == 0 {
		logger.Debug("agents tasks monitor: no tasks to process")
	}
	sort.Slice(tasks, func(i, j int) bool { // process least recently updated tasks first
		return tasks[i].UpdatedOn.Before(tasks[j].UpdatedOn)
	})
	numTasksPerWorker := int(math.Ceil(float64(numTasks) / float64(numTaskProcWorkers)))
	j := 0
	for i := 1; i <= numTaskProcWorkers; i++ {
		if j >= numTasks {
			break
		}
		var tasksForWorker []*agents.Task
		k := j + numTasksPerWorker
		if k <= numTasks {
			tasksForWorker = tasks[j : k]
		} else {
			tasksForWorker = tasks[j : numTasks]
		}
		wg.Add(1)
		go m._processTasks(i, tasksForWorker, wg)
		j = k
	}
}

// start processing tasks and task responses each 10 seconds
func (m *agentEndpointsManager) agentTasksMonitor(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	m.processTasks(wg)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
			case <- ticker.C:
				m.processTasks(wg)
			case <- ctx.Done():
				logger.Info("stopping agent tasks monitor")
				return
		}
	}
}

func initAgentsBackend(r *mux.Router, manager *authManager, ctx context.Context, wg *sync.WaitGroup) {
	agentsBasePath := fmt.Sprintf("/%s", submitws.Agents)
	agentsRouter := r.PathPrefix(agentsBasePath).Subrouter()
	agentsRouter.HandleFunc(fmt.Sprintf("/%s", endpoint), agentEndpoints.agentsEndpoint).Methods(http.MethodGet)
	manager.addPathToMap(fmt.Sprintf("%s/%s", agentsBasePath, endpoint), func (user *users.User, _ string) bool {
		return user.Roles.Contains(users.Agent) || user.Roles.Contains(users.Admin)
	})
	agentsRouter.HandleFunc("/", handleGetAgents).Methods(http.MethodGet)
	manager.addPathToMap(fmt.Sprintf("%s/", agentsBasePath), func (user *users.User, _ string) bool {
		return user.Roles.Contains(users.Admin)
	})
	agentsRouter.HandleFunc(fmt.Sprintf("/{%s}", agentId), handleGetAgent).Methods(http.MethodGet)
	manager.addRegex(regexp.MustCompile(fmt.Sprintf("%s/.", agentsBasePath)), func (user *users.User, _ string) bool {
		return user.Roles.Contains(users.Admin)
	})
	tasksBasePath := fmt.Sprintf("/%s", db.Tasks)
	tasksRouter := r.PathPrefix(tasksBasePath).Subrouter()
	tasksRouter.HandleFunc("/", handleGetTasks).Methods(http.MethodGet)
	manager.addPathToMap(fmt.Sprintf("%s/", tasksBasePath), func (user *users.User, _ string) bool {
		return user.Roles.Contains(users.Admin)
	})
	tasksRouter.HandleFunc(fmt.Sprintf("/{%s}", taskId), handleGetTask).Methods(http.MethodGet)
	manager.addRegex(regexp.MustCompile(fmt.Sprintf("%s/.", tasksBasePath)), func (user *users.User, _ string) bool {
		return user.Roles.Contains(users.Admin)
	})
	taskResponsesBasePath := fmt.Sprintf("/%s", db.TaskResponses)
	taskResponsesRouter := r.PathPrefix(taskResponsesBasePath).Subrouter()
	taskResponsesRouter.HandleFunc("/", handleGetTaskResponses).Methods(http.MethodGet)
	manager.addPathToMap(fmt.Sprintf("%s/", taskResponsesBasePath), func (user *users.User, _ string) bool {
		return user.Roles.Contains(users.Admin)
	})
	taskResponsesRouter.HandleFunc(fmt.Sprintf("/{%s}", taskRespId), handleGetTaskResponse).Methods(http.MethodGet)
	manager.addRegex(regexp.MustCompile(fmt.Sprintf("%s/.", taskResponsesBasePath)), func (user *users.User, _ string) bool {
		return user.Roles.Contains(users.Admin)
	})
	wg.Add(2)
	go agentEndpoints.agentStatusMonitor(ctx, wg)
	go agentEndpoints.agentTasksMonitor(ctx, wg)
}

func init() {
	agentEndpoints = newAgentEndpointsManager()
}
