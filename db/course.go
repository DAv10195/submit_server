package db

import (
	"github.com/DAv10195/submit_server/util/containers"
)

// course struct
type Course struct {
	ABucketElement
	ID			string               	`json:"id"`
	Name		string                 	`json:"name"`
	Students	*containers.StringSet  	`json:"students"`
	Staff		*containers.StringSet 	`json:"staff"`
}

func (c *Course) Key() []byte {
	return []byte(c.ID)
}

func (c *Course) Bucket() []byte {
	return []byte(Courses)
}
