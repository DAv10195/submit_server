package db

import "github.com/DAv10195/submit_server/util"

// course struct
type Course struct {
	ID			string			`json:"id"`
	Name		string			`json:"name"`
	Students	*util.StringSet	`json:"students"`
	Staff		*util.StringSet	`json:"staff"`
}

func (c *Course) Key() []byte {
	return []byte(c.ID)
}

func (c *Course) Bucket() []byte {
	return []byte(Courses)
}
