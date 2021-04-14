package tests

import (
	"fmt"
	"github.com/DAv10195/submit_commons/containers"
	"github.com/DAv10195/submit_server/db"
)

// possible test state values
const (
	Draft 		= iota
	InReview	= iota
	Published	= iota
)

// possible test runs on values
const (
	OnSubmit 	= iota
	OnDemand	= iota
)

// test
type Test struct {
	db.ABucketElement
	Name			string
	State			int
	Files			*containers.StringSet
	AssignmentDef	string
	RunsOn			int
	MessageBox		string
}

func (t *Test) Key() []byte {
	return []byte(fmt.Sprintf("%s%s%s", t.AssignmentDef, db.KeySeparator, t.Name))
}

func (t *Test) Bucket() []byte {
	return []byte(db.Tests)
}
