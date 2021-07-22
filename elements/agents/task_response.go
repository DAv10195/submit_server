package agents

import (
	"encoding/json"
	"github.com/DAv10195/submit_server/db"
)

// task response
type TaskResponse struct {
	db.ABucketElement
	ID			string						`json:"id"`
	Payload		string						`json:"payload"`
	Handler		string						`json:"handler"`
	Task		string						`json:"task"`
	ExecStatus	int							`json:"status"`
	Labels		map[string]interface{}		`json:"labels"`
}

func (t *TaskResponse) Key() []byte {
	return []byte(t.ID)
}

func (t *TaskResponse) Bucket() []byte {
	return []byte(db.TaskResponses)
}

func GetTaskResponse(id string) (*TaskResponse, error) {
	respBytes, err := db.GetFromBucket([]byte(db.TaskResponses), []byte(id))
	if err != nil {
		return nil, err
	}
	resp := &TaskResponse{}
	if err := json.Unmarshal(respBytes, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
