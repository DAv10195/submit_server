package assignments

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DAv10195/submit_commons/containers"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/tests"
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
	Name		string
	DueBy		time.Time
	Course		string
	State		int
	Files		*containers.StringSet
}

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
func DeleteDef(ass *AssignmentDef) error {
	// TODO: delete files in file server
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
		if err := DeleteInstance(inst); err != nil {
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
		if err := tests.Delete(test); err != nil {
			return err
		}
	}
	return db.Delete(ass)
}

func NewDef(course string, dueBy time.Time, name string, asUser string, withDbUpdate bool) (*AssignmentDef, error) {
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
	// TODO: create diretory for ass def in file server
	ass := &AssignmentDef{Course: course, DueBy: dueBy, Name: name, State: Draft, Files: containers.NewStringSet()}
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
