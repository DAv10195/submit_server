package server

type agentTaskResponseHandler func([]byte, map[string]interface{}) error

var agentTaskRespHandlers = make(map[string]agentTaskResponseHandler)
