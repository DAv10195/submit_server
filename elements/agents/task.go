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
	TaskStatusInProgress	= iota
	TaskStatusDone			= iota
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
}

func (t *Task) Key() []byte {
	return []byte(t.ID)
}

func (t *Task) Bucket() []byte {
	return []byte(db.Tasks)
}

func (t *Task) GetWsMessage() (*submitws.Message, error) {
	taskMsgPayload := submitws.Task{}
	taskMsgPayload.ID = t.ID
	taskMsgPayload.Command = t.Command
	taskMsgPayload.Timeout = t.ExecTimeout
	taskMsgPayload.Dependencies = t.Dependencies
	taskMsgPayload.ResponseHandler = t.ResponseHandler
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

type TaskBuilder struct {
	OsType			string
	Architecture	string
	Command			string
	ResponseHandler	string
	ExecTimeout		int
	Dependencies	*containers.StringSet
	Agent			string
	asUser			string
	withDbUpdate	bool
}

func NewTaskBuilder(asUser string, withDbUpdate bool) *TaskBuilder {
	return &TaskBuilder{asUser: asUser, withDbUpdate: withDbUpdate, ExecTimeout: -1, Dependencies: containers.NewStringSet()}
}

func (b *TaskBuilder) WithOsType(osType string) *TaskBuilder {
	b.OsType = osType
	return b
}

func (b *TaskBuilder) WithArchitecture(architecture string) *TaskBuilder {
	b.Architecture = architecture
	return b
}

func (b *TaskBuilder) WithCommand(cmd string) *TaskBuilder {
	b.Command = cmd
	return b
}

func (b *TaskBuilder) WithResponseHandler(respHandler string) *TaskBuilder {
	b.ResponseHandler = respHandler
	return b
}

func (b *TaskBuilder) WithExecTimeout(timeout int) *TaskBuilder {
	b.ExecTimeout = timeout
	return b
}

func (b *TaskBuilder) WithDependencies(dependencies ...string) *TaskBuilder {
	b.Dependencies.Add(dependencies...)
	return b
}

func (b *TaskBuilder) WithAgent(agentId string) *TaskBuilder {
	b.Agent = agentId
	return b
}

func (b *TaskBuilder) Build() (*Task, error) {
	if b.Command == "" {
		return nil, &submiterr.ErrInsufficientData{Message: "task can't have an empty command"}
	}
	if b.ResponseHandler == "" {
		return nil, &submiterr.ErrInsufficientData{Message: "task can't have an empty response handler"}
	}
	if b.ExecTimeout < 0 {
		return nil,  &submiterr.ErrInsufficientData{Message: "task must have a non-negative timeout (seconds) value"}
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
	}
	if b.withDbUpdate {
		if err := db.Update(b.asUser, task); err != nil {
			return nil, err
		}
	}
	return task, nil
}
