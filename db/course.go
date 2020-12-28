package db

import (
	"github.com/DAv10195/submit_server/util/stringset"
)

// course struct
type Course struct {
	ABucketElement
	ID			string              	`json:"id"`
	Name		string                	`json:"name"`
	Students	*stringset.StringSet  	`json:"students"`
	Staff		*stringset.StringSet 	`json:"staff"`
}

func (c *Course) Key() []byte {
	return []byte(c.ID)
}

func (c *Course) Bucket() []byte {
	return []byte(Courses)
}
