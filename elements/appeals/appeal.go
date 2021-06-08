package appeals

import (
	"encoding/json"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/messages"
)

// possible appeal state values
const (
	Open 	= iota
	Closed 	= iota
)

// appeal
type Appeal struct {
	db.ABucketElement
	AssignmentInstance	string
	State				int
	MessageBox			string
}

func Get(id string) (*Appeal, error) {
	appealBytes, err := db.GetFromBucket([]byte(db.Appeals), []byte(id))
	if err != nil {
		return nil, err
	}
	appeal := &Appeal{}
	if err := json.Unmarshal(appealBytes, appeal); err != nil {
		return nil, err
	}
	return appeal, nil
}

// delete the appeal and the message box assigned to it
func Delete(appeal *Appeal) error {
	box, err := messages.Get(appeal.MessageBox)
	if err != nil {
		return err
	}
	if err := messages.Delete(box); err != nil {
		return err
	}
	return db.Delete(appeal)
}

func (a *Appeal) Key() []byte {
	return []byte(a.AssignmentInstance)
}

func (a *Appeal) Bucket() []byte {
	return []byte(db.Appeals)
}
