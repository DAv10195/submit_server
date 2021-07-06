package server

type agentTaskResponseHandler func([]byte, map[string]interface{}) error

var agentTaskRespHandlers = make(map[string]agentTaskResponseHandler)

func handleOnDemandTask(_ []byte, _ map[string]interface{}) error {
	return nil
}

func init() {
	agentTaskRespHandlers[onDemandTask] = handleOnDemandTask
}
