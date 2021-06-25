package messages

import "github.com/DAv10195/submit_server/db"

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
