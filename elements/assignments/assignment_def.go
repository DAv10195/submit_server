package assignments

import (
	"fmt"
	"github.com/DAv10195/submit_commons/containers"
	"github.com/DAv10195/submit_server/db"
	"time"
)

// possible assignment definition state values
const (
	Draft 		= iota
	Published	= iota
	Complete	= iota
)

// assignment definition
type AssignmentDef struct {
	db.ABucketElement
	Name		string
	DueBy		time.Time
	Course		string
	State		int
	Files		*containers.StringSet
}

func (a *AssignmentDef) Key() []byte {
	return []byte(fmt.Sprintf("%s%s%s", a.Course, db.KeySeparator, a.Name))
}

func (a *AssignmentDef) Bucket() []byte {
	return []byte(db.AssignmentDefinitions)
}
