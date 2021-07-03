package tests

import (
	"encoding/json"
	"fmt"
	"github.com/DAv10195/submit_commons/containers"
	submithttp "github.com/DAv10195/submit_commons/http"
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

func New(asUser, assDef, name string, runsOn int, withDbUpdate, withFsUpdate bool) (*Test, error) {
	exists, err := db.KeyExistsInBucket([]byte(db.AssignmentDefinitions), []byte(assDef))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &db.ErrKeyNotFoundInBucket{Bucket: db.AssignmentDefinitions, Key: assDef}
	}
	testKey := fmt.Sprintf("%s%s%s", assDef, db.KeySeparator, name)
	exists, err = db.KeyExistsInBucket([]byte(db.Tests), []byte(testKey))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &db.ErrKeyExistsInBucket{Bucket: db.Tests, Key: testKey}
	}
	if runsOn != OnDemand && runsOn != OnSubmit {
		return nil, fmt.Errorf("invalid runs on value given for test creation ('%d')", runsOn)
	}
	if withFsUpdate {
		split := strings.Split(assDef, db.KeySeparator)
		if len(split) != 3 {
			return nil, fmt.Errorf("invalid assignment def key ('%s')", assDef)
		}
		if err := fs.GetClient().UploadTextToFS(strings.Join([]string{db.Courses, split[0], split[1], split[2], "tests", name, submithttp.FsPlaceHolderFileName}, "/"), []byte("")); err != nil {
			return nil, err
		}
	}
	test := &Test{Name: name, State: Draft, Files: containers.NewStringSet(), AssignmentDef: assDef, RunsOn: runsOn}
	if withDbUpdate {
		msgBox := messages.NewMessageBox()
		test.MessageBox = msgBox.ID
		if err := db.Update(asUser, msgBox, test); err != nil {
			return nil, err
		}
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
