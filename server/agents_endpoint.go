package server

import (
	commons "github.com/DAv10195/submit_commons"
	submitws "github.com/DAv10195/submit_commons/websocket"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

var agentEndpoints *agentEndpointsManager

// a websocket endpoint representing a connection to an agent
type agentEndpoint struct {
	id       string
	conn     *websocket.Conn
	mutex    *sync.RWMutex
	isClosed bool
}

// create a new agent endpoint
func newAgentEndpoint(id string, conn *websocket.Conn) *agentEndpoint {
	return &agentEndpoint{id, conn, &sync.RWMutex{}, false}
}

// read incoming messages from an agent. This function should be called only by a single goroutine
func (e *agentEndpoint) readLoop() {
	for {
		wsMsgType, payload, err := e.conn.ReadMessage()
		if err != nil {
			logger.WithError(err).Errorf("error reading websocket message from agent with id == %s", e.id)
			e.mutex.Lock()
			if !e.isClosed {
				if err := e.conn.Close(); err != nil {
					logger.WithError(err).Errorf("error closing connection to agent with id == %s after write error: err", e.id, err)
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
		logger.Warnf("write called on a closed connection to agent with id == %s", e.id)
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
		logger.Errorf("invalid agent ID sent to agent endpoint [ %s ] from [ %s ]", agentId, r.RemoteAddr)
		return
	}
	wsUpgrade := websocket.Upgrader{}
	conn, err := wsUpgrade.Upgrade(w, r, nil)
	if err != nil {
		logger.WithError(err).Errorf("error upgrading connection from [ %s ] to websocket", r.RemoteAddr)
		return
	}
	logger.Debugf("successfully upgraded connection from [ %s ] to websocket", r.RemoteAddr)
	endpoint := newAgentEndpoint(agentId, conn)
	m.addEndpoint(endpoint)
	endpoint.readLoop()
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

func init() {
	agentEndpoints = newAgentEndpointsManager()
}
