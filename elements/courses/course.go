package courses

import (
	"encoding/json"
	"fmt"
	submiterr "github.com/DAv10195/submit_commons/errors"
	"github.com/DAv10195/submit_server/db"
	"time"
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

// create a new course with the given number and name
func NewCourse(number int, name string, asUser string, withDbUpdate bool) (*Course, error) {
	if number <= 0 {
		return nil, &submiterr.ErrInsufficientData{Message: "number of course must be positive"}
	}
	if name == "" {
		return nil, &submiterr.ErrInsufficientData{Message: "name of course must not be empty"}
	}
	year := time.Now().UTC().Year()
	courseKey := fmt.Sprintf("%d%s%d", number, db.KeySeparator, year)
	exists, err := db.KeyExistsInBucket([]byte(db.Courses), []byte(courseKey))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &db.ErrKeyExistsInBucket{Bucket: db.Courses, Key: courseKey}
	}
	course := &Course{Number: number, Year: year, Name: name}
	if withDbUpdate {
		if err := db.Update(asUser, course); err != nil {
			return nil, err
		}
	}
	return course, nil
}

// return the course with the given number and year if it exists
func Get(number, year int) (*Course, error) {
	courseBytes, err := db.GetFromBucket([]byte(db.Courses), []byte(fmt.Sprintf("%d%s%d", number, db.KeySeparator, year)))
	if err != nil {
		return nil, err
	}
	course := &Course{}
	if err := json.Unmarshal(courseBytes, course); err != nil {
		return nil, err
	}
	return course, nil
}
