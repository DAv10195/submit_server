package courses

import (
	"encoding/json"
	"fmt"
	"github.com/DAv10195/submit_commons/containers"
	submiterr "github.com/DAv10195/submit_commons/errors"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/assignments"
	"time"
)

// course
type Course struct {
	db.ABucketElement
	Number          		int						`json:"number"`
	Year        			int                		`json:"year"`
	Name            		string                	`json:"name"`
	Files					*containers.StringSet	`json:"files"`
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
	// TODO: create diretory for course in file server
	course := &Course{Number: number, Year: year, Name: name, Files: containers.NewStringSet()}
	if withDbUpdate {
		if err := db.Update(asUser, course); err != nil {
			return nil, err
		}
	}
	return course, nil
}

// delete the course and the assignment definitions
func Delete(course *Course) error {
	// TODO: delete files in file server
	var defsToDel []*assignments.AssignmentDef
	if err := db.QueryBucket([]byte(db.AssignmentDefinitions), func(_, elemBytes []byte) error {
		def := &assignments.AssignmentDef{}
		if err := json.Unmarshal(elemBytes, def); err != nil {
			return err
		}
		if def.Course == string(course.Key()) {
			defsToDel = append(defsToDel, def)
		}
		return nil
	}); err != nil {
		return err
	}
	for _, def := range defsToDel {
		if err := assignments.DeleteDef(def); err != nil {
			return err
		}
	}
	return db.Delete(course)
}

// return the course with the given number and year if it exists
func Get(id string) (*Course, error) {
	courseBytes, err := db.GetFromBucket([]byte(db.Courses), []byte(id))
	if err != nil {
		return nil, err
	}
	course := &Course{}
	if err := json.Unmarshal(courseBytes, course); err != nil {
		return nil, err
	}
	return course, nil
}
