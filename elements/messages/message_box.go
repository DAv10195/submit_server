package messages

import (
	"encoding/json"
	commons "github.com/DAv10195/submit_commons"
	"github.com/DAv10195/submit_commons/containers"
	"github.com/DAv10195/submit_server/db"
)

// message box
type MessageBox struct {
	db.ABucketElement
	ID			string					`json:"id"`
	Messages	*containers.StringSet	`json:"messages"`
}

func (m *MessageBox) Key() []byte {
	return []byte(m.ID)
}

func (m *MessageBox) Bucket() []byte {
	return []byte(db.MessageBoxes)
}

// create a new message box (no update in db)
func NewMessageBox() *MessageBox {
	return &MessageBox{
		ID: commons.GenerateUniqueId(),
		Messages: containers.NewStringSet(),
	}
}

// get message box by id
func Get(id string) (*MessageBox, error) {
	boxBytes, err := db.GetFromBucket([]byte(db.MessageBoxes), []byte(id))
	if err != nil {
		return nil, err
	}
	box := &MessageBox{}
	if err := json.Unmarshal(boxBytes, box); err != nil {
		return nil, err
	}
	return box, nil
}

// delete a message box and all of the messages associated with it
func Delete(box *MessageBox) error {
	var messagesToDel [][]byte
	for _, msgKey := range box.Messages.Slice() {
		messagesToDel = append(messagesToDel, []byte(msgKey))
	}
	if err := db.DeleteKeysFromBucket([]byte(db.Messages), messagesToDel...); err != nil {
		return err
	}
	return db.Delete(box)
}
