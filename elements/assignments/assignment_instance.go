package assignments

import (
	"encoding/json"
	"fmt"
	"github.com/DAv10195/submit_commons/containers"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/appeals"
	"time"
)

// possible assignment instance state values
const (
	Assigned 	= iota
	Submitted	= iota
	Graded		= iota
)

// assignment instance
type AssignmentInstance struct {
	db.ABucketElement
	UserName		string
	AssignmentDef 	string
	State			int
	Files			*containers.StringSet
	DueBy			time.Time
	MarkedAsCopy	bool
	Grade			int
}

func GetInstance(id string) (*AssignmentInstance, error) {
	assBytes, err := db.GetFromBucket([]byte(db.AssignmentInstances), []byte(id))
	if err != nil {
		return nil, err
	}
	ass := &AssignmentInstance{}
	if err := json.Unmarshal(assBytes, ass); err != nil {
		return nil, err
	}
	return ass, nil
}

// delete the assignment instance, appeals associated and the files associated with it
func DeleteInstance(ass *AssignmentInstance) error {
	// TODO: delete files in file server
	appeal, err := appeals.Get(string(ass.Key()))
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); !ok {
			return err
		}
	}
	if appeal != nil {
		if err := appeals.Delete(appeal); err != nil {
			return err
		}
	}
	return db.Delete(ass)
}

func (a *AssignmentInstance) Key() []byte {
	return []byte(fmt.Sprintf("%s%s%s", a.AssignmentDef, db.KeySeparator, a.UserName))
}

func (a *AssignmentInstance) Bucket() []byte {
	return []byte(db.AssignmentInstances)
}
