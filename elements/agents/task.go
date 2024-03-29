package agents

import (
	"encoding/json"
	commons "github.com/DAv10195/submit_commons"
	"github.com/DAv10195/submit_commons/containers"
	submiterr "github.com/DAv10195/submit_commons/errors"
	submitws "github.com/DAv10195/submit_commons/websocket"
	"github.com/DAv10195/submit_server/db"
)

// possible status values
const (
	TaskStatusReady			= iota
	TaskStatusDone			= iota
	TaskStatusAssigned		= iota
	TaskStatusInProgress	= iota
	TaskStatusProcessing	= iota
	TaskStatusOk			= iota
	TaskStatusTimeout		= iota
	TaskStatusError			= iota
)

// task
type Task struct {
	db.ABucketElement
	ID				string					`json:"id"`
	OsType			string					`json:"os_type"`
	Architecture	string					`json:"architecture"`
	Command			string					`json:"command"`
	ResponseHandler	string					`json:"response_handler"`
	ExecTimeout		int						`json:"timeout"`
	TaskResponse	string					`json:"task_response"`
	Dependencies	*containers.StringSet	`json:"dependencies"`
	Status			int						`json:"status"`
	Description		string					`json:"description"`
	Agent			string					`json:"agent"`
	Labels			map[string]interface{}	`json:"labels"`
}

func (t *Task) Key() []byte {
	return []byte(t.ID)
}

func (t *Task) Bucket() []byte {
	return []byte(db.Tasks)
}

// return a websocket message representing the task
func (t *Task) GetWsMessage() (*submitws.Message, error) {
	taskMsgPayload := submitws.Task{}
	taskMsgPayload.ID = t.ID
	taskMsgPayload.Command = t.Command
	taskMsgPayload.Timeout = t.ExecTimeout
	taskMsgPayload.Dependencies = t.Dependencies
	taskMsgPayload.ResponseHandler = t.ResponseHandler
	taskMsgPayload.Labels = t.Labels
	payload, err := json.Marshal(taskMsgPayload)
	if err != nil {
		return nil, err
	}
	msg, err := submitws.NewMessage(submitws.MessageTypeTask, payload)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// get a task by id
func GetTask(id string) (*Task, error) {
	taskBytes, err := db.GetFromBucket([]byte(db.Tasks), []byte(id))
	if err != nil {
		return nil, err
	}
	task := &Task{}
	if err := json.Unmarshal(taskBytes, task); err != nil {
		return nil, err
	}
	return task, nil
}

// builds tasks while performing the required validations
type TaskBuilder struct {
	OsType			string
	Architecture	string
	Command			string
	ResponseHandler	string
	ExecTimeout		int
	Dependencies	*containers.StringSet
	Agent			string
	Labels			map[string]interface{}
	asUser			string
	withDbUpdate	bool
}

// returns a new instance of TaskBuilder
func NewTaskBuilder(asUser string, withDbUpdate bool) *TaskBuilder {
	return &TaskBuilder{asUser: asUser, withDbUpdate: withDbUpdate, Dependencies: containers.NewStringSet(), Labels: make(map[string]interface{})}
}

// set os type to run task on
func (b *TaskBuilder) WithOsType(osType string) *TaskBuilder {
	b.OsType = osType
	return b
}

// set architecture to run task on
func (b *TaskBuilder) WithArchitecture(architecture string) *TaskBuilder {
	b.Architecture = architecture
	return b
}

// set command to run
func (b *TaskBuilder) WithCommand(cmd string) *TaskBuilder {
	b.Command = cmd
	return b
}

// set the handler of the response returned after the execution
func (b *TaskBuilder) WithResponseHandler(respHandler string) *TaskBuilder {
	b.ResponseHandler = respHandler
	return b
}

// set execution timeout
func (b *TaskBuilder) WithExecTimeout(timeout int) *TaskBuilder {
	b.ExecTimeout = timeout
	return b
}

// add dependencies paths to download from the submit file server
func (b *TaskBuilder) WithDependencies(dependencies ...string) *TaskBuilder {
	b.Dependencies.Add(dependencies...)
	return b
}

// set a specific agent id to run the task
func (b *TaskBuilder) WithAgent(agentId string) *TaskBuilder {
	b.Agent = agentId
	return b
}

// add a label to the task
func (b *TaskBuilder) WithLabel(name string, value interface{}) *TaskBuilder {
	b.Labels[name] = value
	return b
}

// build the task with the parameters set
func (b *TaskBuilder) Build() (*Task, error) {
	if b.Command == "" {
		return nil, &submiterr.ErrInsufficientData{Message: "task can't have an empty command"}
	}
	if b.ResponseHandler == "" {
		return nil, &submiterr.ErrInsufficientData{Message: "task can't have an empty response handler"}
	}
	if b.ExecTimeout <= 0 {
		return nil,  &submiterr.ErrInsufficientData{Message: "task must have a positive timeout (seconds) value"}
	}
	task := &Task{
		ID: commons.GenerateUniqueId(),
		OsType: b.OsType,
		Architecture: b.Architecture,
		Command: b.Command,
		ResponseHandler: b.ResponseHandler,
		ExecTimeout: b.ExecTimeout,
		Dependencies: b.Dependencies,
		Status: TaskStatusReady,
		Agent: b.Agent,
		Labels: b.Labels,
	}
	if b.withDbUpdate {
		if err := db.Update(b.asUser, task); err != nil {
			return nil, err
		}
	}
	return task, nil
}
