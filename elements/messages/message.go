package messages

import (
	commons "github.com/DAv10195/submit_commons"
	"github.com/DAv10195/submit_server/db"
)

// message
type Message struct {
	db.ABucketElement
	ID		string	`json:"id"`
	From	string	`json:"from"`
	Text	string	`json:"text"`
}

func (m *Message) Key() []byte {
	return []byte(m.ID)
}

func (m *Message) Bucket() []byte {
	return []byte(db.Messages)
}

// create a new message in the box with the given id
func NewMessage(from, text, boxId string, withDbUpdate bool) (*Message, *MessageBox, error) {
	msgBox, err := Get(boxId)
	if err != nil {
		return nil, nil, err
	}
	msg := &Message{From: from, Text: text, ID: commons.GenerateUniqueId()}
	if withDbUpdate {
		msgBox.Messages.Add(msg.ID)
		if err := db.Update(from, msg, msgBox); err != nil {
			return nil, nil, err
		}
	}
	return msg, msgBox, nil
}
