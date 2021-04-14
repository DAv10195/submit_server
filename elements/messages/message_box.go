package messages

import (
	commons "github.com/DAv10195/submit_commons"
	"github.com/DAv10195/submit_commons/containers"
	"github.com/DAv10195/submit_server/db"
)

// message box
type MessageBox struct {
	db.ABucketElement
	ID			string
	Messages	*containers.StringSet
}

func (m *MessageBox) Key() []byte {
	return []byte(m.ID)
}

func (m *MessageBox) Bucket() []byte {
	return []byte(db.MessageBoxes)
}

func NewMessageBox() *MessageBox {
	return &MessageBox{
		ID: commons.GenerateUniqueId(),
		Messages: containers.NewStringSet(),
	}
}
