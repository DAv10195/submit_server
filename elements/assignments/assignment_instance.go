package assignments

import (
	"fmt"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/util/containers"
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

func (a *AssignmentInstance) Key() []byte {
	return []byte(fmt.Sprintf("%s%s%s", a.AssignmentDef, db.KeySeparator, a.UserName))
}

func (a *AssignmentInstance) Bucket() []byte {
	return []byte(db.AssignmentInstances)
}
