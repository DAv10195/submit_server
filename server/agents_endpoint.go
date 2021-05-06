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
	"net/http"
	"regexp"
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
			logger.WithError(err).Errorf("error closing connection to agent with id == %s after write error: err", e.id, err)
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
	logger.Debug("agents status monitor: processing agents keepalive messages...")
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
	logger.Debug("agents status monitor: finished processing agents keepalive messages")
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
				logger.Debug("stopping agent status monitor")
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
	agentsRouter.HandleFunc(fmt.Sprintf("/{%s}/get", agentId), handleGetAgent).Methods(http.MethodGet)
	manager.addRegex(regexp.MustCompile(fmt.Sprintf("%s/.", agentsBasePath)), func (user *users.User, _ string) bool {
		return user.Roles.Contains(users.Admin)
	})
	wg.Add(1)
	go agentEndpoints.agentStatusMonitor(ctx, wg)
}

func init() {
	agentEndpoints = newAgentEndpointsManager()
}
