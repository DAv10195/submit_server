package assignments

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DAv10195/submit_commons/containers"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/appeals"
	"github.com/DAv10195/submit_server/fs"
	"strings"
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
	UserName		string					`json:"user_name"`
	AssignmentDef 	string					`json:"assignment_def"`
	State			int						`json:"state"`
	Files			*containers.StringSet	`json:"files"`
	DueBy			time.Time				`json:"due_by"`
	MarkedAsCopy	bool					`json:"copy"`
	Grade			int						`json:"grade"`
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
func DeleteInstance(ass *AssignmentInstance, withFsUpdate bool) error {
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
	if err := db.Delete(ass); err != nil {
		return err
	}
	if withFsUpdate {
		split := strings.Split(ass.AssignmentDef, db.KeySeparator)
		if len(split) != 3 {
			fmt.Errorf("invalid assignment def key ('%s')", ass.AssignmentDef)
		}
		if err := fs.GetClient().Delete(strings.Join([]string{db.Courses, split[0], split[1], split[2], ass.UserName}, "/")); err != nil {
			return err
		}
	}
	return nil
}

func NewInstance(course string, dueBy time.Time, assName string, userName string, asUser string, withDbUpdate bool, withFsUpdate bool) (*AssignmentInstance, error) {
	assDefKey := fmt.Sprintf("%s%s%s", course, db.KeySeparator, assName)
	exists, err := db.KeyExistsInBucket([]byte(db.AssignmentDefinitions), []byte(assDefKey))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &db.ErrKeyNotFoundInBucket{Bucket: db.AssignmentDefinitions, Key: assDefKey}
	}
	assInstKey := fmt.Sprintf("%s%s%s", assDefKey, db.KeySeparator, userName)
	exists, err = db.KeyExistsInBucket([]byte(db.AssignmentInstances), []byte(assInstKey))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &db.ErrKeyExistsInBucket{Bucket: db.AssignmentInstances, Key: assInstKey}
	}
	if dueBy.Before(time.Now().UTC()) {
		return nil, errors.New("given due by time is before current UTC time")
	}
	if withFsUpdate {
		split := strings.Split(course, db.KeySeparator)
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid course key ('%s')", course)
		}
		if err := fs.GetClient().UploadTextToFS(strings.Join([]string{db.Courses, split[0], split[1], assName, userName, submithttp.FsPlaceHolderFileName}, "/"), []byte("")); err != nil {
			return nil, err
		}
	}
	ass := &AssignmentInstance{UserName: userName, AssignmentDef: assDefKey, State: Assigned, DueBy: dueBy}
	if withDbUpdate {
		if err := db.Update(asUser, ass); err != nil {
			return nil, err
		}
	}
	return ass, nil
}

func (a *AssignmentInstance) Key() []byte {
	return []byte(fmt.Sprintf("%s%s%s", a.AssignmentDef, db.KeySeparator, a.UserName))
}

func (a *AssignmentInstance) Bucket() []byte {
	return []byte(db.AssignmentInstances)
}
