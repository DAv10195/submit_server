package tests

import (
	"encoding/json"
	"fmt"
	"github.com/DAv10195/submit_commons/containers"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/messages"
	"github.com/DAv10195/submit_server/fs"
	"strings"
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
	Name			string					`json:"name"`
	State			int						`json:"state"`
	Files			*containers.StringSet	`json:"files"`
	AssignmentDef	string					`json:"assignment_def"`
	RunsOn			int						`json:"runs_on"`
	MessageBox		string					`json:"message_box"`
}

func (t *Test) Key() []byte {
	return []byte(fmt.Sprintf("%s%s%s", t.AssignmentDef, db.KeySeparator, t.Name))
}

func (t *Test) Bucket() []byte {
	return []byte(db.Tests)
}

func Get(id string) (*Test, error) {
	testyBytes, err := db.GetFromBucket([]byte(db.Tests), []byte(id))
	if err != nil {
		return nil, err
	}
	test := &Test{}
	if err := json.Unmarshal(testyBytes, test); err != nil {
		return nil, err
	}
	return test, nil
}

func Delete(t *Test, withFsUpdate bool) error {
	box, err := messages.Get(t.MessageBox)
	if err != nil {
		return err
	}
	if err := messages.Delete(box); err != nil {
		return err
	}
	if err := db.Delete(t); err != nil {
		return err
	}
	if withFsUpdate {
		split := strings.Split(t.AssignmentDef, db.KeySeparator)
		if len(split) != 3 {
			return fmt.Errorf("invalid assignment def key ('%s')", t.AssignmentDef)
		}
		if err := fs.GetClient().Delete(strings.Join([]string{db.Courses, split[0], split[1], split[2], "tests"}, "/")); err != nil {
			return err
		}
	}
	return nil
}
