package courses

import (
	"fmt"
	"github.com/DAv10195/submit_server/db"
)

// course
type Course struct {
	db.ABucketElement
	Number          		int						`json:"number"`
	Year        			int                		`json:"year"`
	Name            		string                	`json:"name"`
}

func (c *Course) Key() []byte {
	return []byte(fmt.Sprintf("%d%s%d", c.Number, db.KeySeparator, c.Year))
}

func (c *Course) Bucket() []byte {
	return []byte(db.Courses)
}
