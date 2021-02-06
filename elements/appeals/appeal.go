package appeals

import "github.com/DAv10195/submit_server/db"

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

func (a *Appeal) Key() []byte {
	return []byte(a.AssignmentInstance)
}

func (a *Appeal) Bucket() []byte {
	return []byte(db.Appeals)
}
