package server

import (
	"fmt"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/util/containers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
	"sync"
	"time"
)

var agentMessageTypes = containers.NewStringSet()

// agent message
type agentMessage struct {
	msgType		string
	payload		[]byte
}

// an interface to be implemented by all handlers that wish to register to agent message events
type agentMessageHandler interface {
	messageTypes() []string
	handle(*agentSession, *agentMessage)
}

// agent session
type agentSession struct {
	mutex			*sync.Mutex
	conn 			*websocket.Conn
	wg				*sync.WaitGroup
	lastKeepalive	time.Time
	createdAt		time.Time
	messageHandlers	map[string][]agentMessageHandler
}

func (s *agentSession) readLoop() {
	s.wg.Add(1)
	defer s.wg.Done()
	for {
		wsProtocolMsgType, payload, err := s.conn.ReadMessage()
		if err != nil {
			logger.WithError(err).Errorf("error reading ws message from agent on %v", s.conn.RemoteAddr())
			return
		}
		if wsProtocolMsgType != websocket.BinaryMessage {
			logger.Errorf("invalid message type received from agent [ %d ]. Expected type is %d", wsProtocolMsgType, websocket.BinaryMessage)
			continue
		}
		payloadStr := string(payload)
		newLineSepIndex := strings.IndexRune(payloadStr, '\n')
		if newLineSepIndex < 0 {
			logger.Errorf("invalid message received from agent on %v: couldn't resolve message type", s.conn.RemoteAddr())
			continue
		}
		msgType, payloadStr := payloadStr[0 : newLineSepIndex], payloadStr[newLineSepIndex : ]
		msg := &agentMessage{
			msgType: msgType,
			payload: []byte(payloadStr),
		}
		handlers, found := s.messageHandlers[msgType]
		if !found {
			logger.Errorf("invalid message type [ %s ] received from agent on %v", s.conn.RemoteAddr())
			continue
		}
		go func() {
			logger.Debugf("handling message from agent on %v", s.conn.RemoteAddr())
			for _, handler := range handlers {
				handler.handle(s, msg)
			}
		}()
	}
}

func (s *agentSession) write(msg *agentMessage) {
	if err := s.conn.WriteMessage(websocket.BinaryMessage, []byte(fmt.Sprintf("%s\n%s", msg.msgType, msg.payload))); err != nil {
		logger.WithError(err).Errorf("error writing message to agent on %v", s.conn.RemoteAddr())
	}
}

func (s *agentSession) close() {
	if err := s.conn.Close(); err != nil {
		logger.WithError(err).Errorf("error closing ws connection from agent on %v", s.conn.RemoteAddr())
	}
	s.wg.Wait()
}

// a registry to hold information about and manage the connected agents
type agentsRegistry struct {
	mutex			*sync.Mutex
	sessions		map[string]*agentSession
	messageHandlers map[string][]agentMessageHandler
}

func (r *agentsRegistry) addSession(conn *websocket.Conn) {
	r.mutex.Lock()
	r.mutex.Unlock()
	addrStr := conn.RemoteAddr().String()
	if _, found := r.sessions[addrStr]; found {
		logger.Warnf("an agent session with %s already exists. Ignoring call for a new one...", addrStr)
		return
	}
	session := &agentSession{
		mutex:				&sync.Mutex{},
		conn:				conn,
		wg:					&sync.WaitGroup{},
		createdAt:			time.Now().UTC(),
		messageHandlers: 	r.messageHandlers,
	}
	r.sessions[addrStr] = session
	go session.readLoop()
}

func (r *agentsRegistry) addMessageHandler(handler agentMessageHandler) {
	for _, msgType := range handler.messageTypes() {
		if !agentMessageTypes.Contains(msgType) {
			// shouldn't happen, but it is better to panic her then not check this at all
			panic(fmt.Sprintf("invalid agent message type: %s", msgType))
		}
		r.messageHandlers[msgType] = append(r.messageHandlers[msgType], handler)
	}
}

func (r *agentsRegistry) clearExpiredSessions() {
	r.mutex.Lock()
	r.mutex.Unlock()
	for _, session := range r.sessions {
		session.mutex.Lock()
		expired := false
		now := time.Now().UTC()
		if !session.lastKeepalive.IsZero() { // means that at least one keepalive message was already sent by the agent
			expired = session.lastKeepalive.Add(time.Minute).Before(now)
		} else { // no keepalive sent from the agent yet so lets validate that session was created at most a minute ago
			expired = session.createdAt.Add(time.Minute).Before(now)
		}
		if expired {
			session.close()
		}
		session.mutex.Unlock()
	}
}

func (r *agentsRegistry) closeAllSessions() {
	r.mutex.Lock()
	r.mutex.Unlock()
	for _, session := range r.sessions {
		session.mutex.Lock()
		session.close()
		session.mutex.Unlock()
	}
}

func newAgentsRegistry() *agentsRegistry {
	return &agentsRegistry{
		mutex:    			&sync.Mutex{},
		sessions: 			make(map[string]*agentSession),
		messageHandlers: 	make(map[string][]agentMessageHandler),
	}
}

func baseWsHandler(registry *agentsRegistry) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !websocket.IsWebSocketUpgrade(r) {
			writeStrErrResp(w, r, http.StatusBadRequest, "invalid request (not ws upgrade)")
			return
		}
		wsUpgrade := &websocket.Upgrader{}
		conn, err := wsUpgrade.Upgrade(w, r, nil)
		if err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
		logger.Debugf("received ws connection from %v", conn.RemoteAddr())
		registry.addSession(conn)
	}
}

func initAgentsRouter(r *mux.Router, manager *authManager) func() {
	agentsBasePath := fmt.Sprintf("/%s", agents)
	agentsRouter := r.PathPrefix(agentsBasePath).Subrouter()
	registry := newAgentsRegistry()
	initMessageHandlers(registry)
	wg := &sync.WaitGroup{}
	stopChan := make(chan struct{})
	go func() {
		wg.Add(1)
		defer wg.Done()
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
				case <- ticker.C:
					registry.clearExpiredSessions()
				case <- stopChan:
					return
			}
		}
	}()
	agentsRouter.HandleFunc("/ws", baseWsHandler(registry))
	manager.addPathToMap(fmt.Sprintf("%s/ws", agentsBasePath), func (user *users.User, _ string) bool {
		return user.Roles.Contains(users.Agent)
	})
	return func() {
		stopChan <- struct{}{}
		wg.Wait()
		registry.closeAllSessions()
	}
}
