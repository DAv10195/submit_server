package server

import "time"

type keepaliveHandler struct {}

func (h *keepaliveHandler) messageTypes() []string {
	return []string{MessageTypeKeepalive}
}

func (h *keepaliveHandler) handle(sess *agentSession, _ *agentMessage) {
	sess.mutex.Lock()
	defer sess.mutex.Unlock()
	sess.lastKeepalive = time.Now().UTC()
}

func initMessageHandlers(registry *agentsRegistry) {
	registry.addMessageHandler(&keepaliveHandler{})
}
