package courses

import (
	"fmt"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/util/containers"
)

// course struct
type Course struct {
	db.ABucketElement
	ID			string               	`json:"id"`
	Semester	string					`json:"semester"`
	Name		string                 	`json:"name"`
	Students	*containers.StringSet  	`json:"students"`
	Staff		*containers.StringSet 	`json:"staff"`
	Assignments	*containers.StringSet	`json:"assignments"`
}

func (c *Course) Key() []byte {
	return []byte(fmt.Sprintf("%s:%s", c.ID, c.Semester))
}

func (c *Course) Bucket() []byte {
	return []byte(db.Courses)
}
