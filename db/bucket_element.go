package db

import "time"

// implementors of this interface specify to which Bucket they should be put and what should be their unique Key
type IBucketElement interface {
	// get the key that should be associated with this element
	Key() []byte
	// get the name of the bucket that this element should be stored in
	Bucket() []byte
	// mark as created in the DB by the given user
	MarkInsert(user string)
	// mark as updated in the DB by the given user
	MarkUpdate(user string)
}

// a common struct that should be embedded into all implementors of IBucketElement
type ABucketElement struct {
	CreatedBy	string		`json:"created_by"`
	CreatedOn	time.Time	`json:"created_on"`
	UpdatedBy	string		`json:"updated_by"`
	UpdatedOn	time.Time	`json:"updated_on"`
}

func (e *ABucketElement) MarkInsert(user string) {
	e.CreatedBy = user
	e.CreatedOn = time.Now().UTC()
	e.UpdatedBy = e.CreatedBy
	e.UpdatedOn = e.CreatedOn
}

func (e *ABucketElement) MarkUpdate(user string) {
	e.UpdatedBy = user
	e.UpdatedOn = time.Now().UTC()
}
