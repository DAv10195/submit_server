package db

import "github.com/DAv10195/submit_server/util"

// manage all access to Courses
type CourseRegistry struct {
	db *BoltDBManager
}

// get courses with the given ids
func (u *CourseRegistry) Get(ids ...string) ([]*Course, error) {
	var elements []BucketElement
	for _, id := range ids {
		elements = append(elements, &Course{id, "", nil, nil})
	}
	if err := u.db.Get(elements...); err != nil {
		return nil, err
	}
	var users []*Course
	for _, element := range elements {
		users = append(users, element.(*Course))
	}
	return users, nil
}

// update the given courses in the DB
func (u *CourseRegistry) Update(courses ...*Course) error {
	var elements []BucketElement
	for _, course := range courses {
		elements = append(elements, course)
	}
	if err := u.db.Put(elements...); err != nil {
		return err
	}
	return nil
}

// delete the Courses with the given ids from the DB
func (u *CourseRegistry) Delete(ids ...string) error {
	var elements []BucketElement
	for _, id := range ids {
		elements = append(elements, &Course{id, "", nil, nil})
	}
	if err := u.db.Delete(elements...); err != nil {
		return err
	}
	return nil
}

// return a new Course struct
func (u *CourseRegistry) New(id, name string) *Course {
	return &Course{id, name, util.NewStringSet(), util.NewStringSet()}
}
