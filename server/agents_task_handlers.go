package server

type agentTaskResponseHandler func([]byte) error

var agentTaskRespHandlers = make(map[string]agentTaskResponseHandler)
