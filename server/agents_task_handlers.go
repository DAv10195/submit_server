package server

type agentTaskResponseHandler func([]byte, map[string]interface{}) error

var agentTaskRespHandlers = make(map[string]agentTaskResponseHandler)

func handleTask(payload []byte, labels map[string]interface{}) error {
	logger.Infof("payload from task: %s, labels: %v", string(payload), labels)
	return nil
}

func init() {
	agentTaskRespHandlers["handle_task"] = handleTask
}
