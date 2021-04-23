package server

import submitws "github.com/DAv10195/submit_commons/websocket"

type agentMessageHandler func(string, []byte)

var agentMsgHandlers = make(map[string]agentMessageHandler)

func handleKeepalive(agentId string, payload []byte) {
	logger.Infof("received message [ %s ] from agent with id == %s", string(payload), agentId)
}

func init() {
	agentMsgHandlers[submitws.MessageTypeKeepalive] = handleKeepalive
}
