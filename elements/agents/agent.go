package agents

import (
	"encoding/json"
	"github.com/DAv10195/submit_server/db"
	"time"
)

// possible status values
const (
	Up 		= iota
	Down	= iota
)

// agent
type Agent struct {
	db.ABucketElement
	ID				string		`json:"id"`
	User			string		`json:"logged_in_user"`
	Hostname 		string		`json:"hostname"`
	IpAddress		string		`json:"ip_address"`
	OsType			string		`json:"os_type"`
	Architecture	string		`json:"architecture"`
	Status			int			`json:"status"`
	NumRunningTasks	int			`json:"num_running_tasks"`
	LastKeepalive	time.Time	`json:"last_keepalive"`
}

func (a *Agent) Key() []byte {
	return []byte(a.ID)
}

func (a *Agent) Bucket() []byte {
	return []byte(db.Agents)
}

// return the agent represented by the given agent id if that agent exists
func Get(agentId string) (*Agent, error) {
	agentBytes, err := db.GetFromBucket([]byte(db.Agents), []byte(agentId))
	if err != nil {
		return nil, err
	}
	agent := &Agent{}
	if err = json.Unmarshal(agentBytes, agent); err != nil {
		return nil, err
	}

	return agent, nil
}
