package assignments

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DAv10195/submit_commons/containers"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/tests"
	"github.com/DAv10195/submit_server/fs"
	"strings"
	"time"
)

// possible assignment definition state values
const (
	Draft 		= iota
	Published	= iota
)

// assignment definition
type AssignmentDef struct {
	db.ABucketElement
	Name			string					`json:"name"`
	DueBy			time.Time				`json:"due_by"`
	Course			string					`json:"course"`
	State			int						`json:"state"`
	Files			*containers.StringSet	`json:"files"`
	RequiredFiles 	*containers.StringSet	`json:"required_files"`
}

// get ass def by id
func GetDef(id string) (*AssignmentDef, error) {
	assBytes, err := db.GetFromBucket([]byte(db.AssignmentDefinitions), []byte(id))
	if err != nil {
		return nil, err
	}
	ass := &AssignmentDef{}
	if err := json.Unmarshal(assBytes, ass); err != nil {
		return nil, err
	}
	return ass, nil
}

// delete the assignment definition, instances, relevant tests and files
func DeleteDef(ass *AssignmentDef, withFsUpdate bool) error {
	var instToDel []*AssignmentInstance
	if err := db.QueryBucket([]byte(db.AssignmentInstances), func(_, elemBytes []byte) error {
		inst := &AssignmentInstance{}
		if err := json.Unmarshal(elemBytes, inst); err != nil {
			return err
		}
		if inst.AssignmentDef == string(ass.Key()) {
			instToDel = append(instToDel, inst)
		}
		return nil
	}); err != nil {
		return err
	}
	for _, inst := range instToDel {
		if err := DeleteInstance(inst, withFsUpdate); err != nil {
			return err
		}
	}
	var testsToDel []*tests.Test
	if err := db.QueryBucket([]byte(db.Tests), func(_, elemBytes []byte) error {
		test := &tests.Test{}
		if err := json.Unmarshal(elemBytes, test); err != nil {
			return err
		}
		if test.AssignmentDef == string(ass.Key()) {
			testsToDel = append(testsToDel, test)
		}
		return nil
	}); err != nil {
		return err
	}
	for _, test := range testsToDel {
		if err := tests.Delete(test, withFsUpdate); err != nil {
			return err
		}
	}
	if err := db.Delete(ass); err != nil {
		return err
	}
	if withFsUpdate {
		split := strings.Split(ass.Course, db.KeySeparator)
		if len(split) != 2 {
			return fmt.Errorf("invalid course key ('%s')", ass.Course)
		}
		if err := fs.GetClient().Delete(strings.Join([]string{db.Courses, split[0], split[1], ass.Name}, "/")); err != nil {
			return err
		}
	}
	return nil
}

// create new assignment definition
func NewDef(course string, dueBy time.Time, name string, asUser string, withDbUpdate bool, withFsUpdate bool) (*AssignmentDef, error) {
	exists, err := db.KeyExistsInBucket([]byte(db.Courses), []byte(course))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &db.ErrKeyNotFoundInBucket{Bucket: db.Courses, Key: course}
	}
	assKey := fmt.Sprintf("%s%s%s", course, db.KeySeparator, name)
	exists, err = db.KeyExistsInBucket([]byte(db.AssignmentDefinitions), []byte(assKey))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &db.ErrKeyExistsInBucket{Bucket: db.AssignmentDefinitions, Key: assKey}
	}
	if dueBy.Before(time.Now().UTC()) {
		return nil, errors.New("given due by time is before current UTC time")
	}
	if withFsUpdate {
		split := strings.Split(course, db.KeySeparator)
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid course key ('%s')", course)
		}
		// each ass def should also have a test folder, so create the entire hierarchy down to the tests folder and the ass folder will be created on the way
		if err := fs.GetClient().UploadTextToFS(strings.Join([]string{db.Courses, split[0], split[1], name, "tests", submithttp.FsPlaceHolderFileName}, "/"), []byte("")); err != nil {
			return nil, err
		}
	}
	ass := &AssignmentDef{Course: course, DueBy: dueBy, Name: name, State: Draft, Files: containers.NewStringSet(), RequiredFiles: containers.NewStringSet()}
	if withDbUpdate {
		if err := db.Update(asUser, ass); err != nil {
			return nil, err
		}
	}
	return ass, nil
}

func (a *AssignmentDef) Key() []byte {
	return []byte(fmt.Sprintf("%s%s%s", a.Course, db.KeySeparator, a.Name))
}

func (a *AssignmentDef) Bucket() []byte {
	return []byte(db.AssignmentDefinitions)
}
