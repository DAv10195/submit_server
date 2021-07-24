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
	AssignmentInstance	string	`json:"assignment_instance"`
	State				int		`json:"state"`
	MessageBox			string	`json:"message_box"`
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

// create a new appeal
func New(assInst string, asUser string, withDbUpdate bool) (*Appeal, error) {
	exists, err := db.KeyExistsInBucket([]byte(db.AssignmentInstances), []byte(assInst))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &db.ErrKeyNotFoundInBucket{Bucket: db.AssignmentInstances, Key: assInst}
	}
	exists, err = db.KeyExistsInBucket([]byte(db.Appeals), []byte(assInst))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &db.ErrKeyExistsInBucket{Bucket: db.Appeals, Key: assInst}
	}
	appeal := &Appeal{AssignmentInstance: assInst, State: Open}
	if withDbUpdate {
		mBox := messages.NewMessageBox()
		appeal.MessageBox = mBox.ID
		if err := db.Update(asUser, mBox, appeal); err != nil {
			return nil, err
		}
	}
	return appeal, nil
}

func (a *Appeal) Key() []byte {
	return []byte(a.AssignmentInstance)
}

func (a *Appeal) Bucket() []byte {
	return []byte(db.Appeals)
}
