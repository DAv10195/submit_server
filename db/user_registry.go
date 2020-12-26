package db

import "github.com/DAv10195/submit_server/util"

// manage all access to Users
type UserRegistry struct {
	db	*BoltDBManager
}

// get users with the given names
func (u *UserRegistry) Get(names ...string) ([]*User, error) {
	var elements []BucketElement
	for _, name := range names {
		elements = append(elements, &User{name, "", "", nil, nil})
	}
	if err := u.db.Get(elements...); err != nil {
		return nil, err
	}
	var users []*User
	for _, element := range elements {
		users = append(users, element.(*User))
	}
	return users, nil
}

// update the given users in the DB
func (u *UserRegistry) Update(users ...*User) error {
	var elements []BucketElement
	for _, user := range users {
		elements = append(elements, user)
	}
	if err := u.db.Put(elements...); err != nil {
		return err
	}
	return nil
}

// delete the Users with the given names from the DB
func (u *UserRegistry) Delete(names ...string) error {
	var elements []BucketElement
	for _, name := range names {
		elements = append(elements, &User{name, "", "", nil, nil})
	}
	if err := u.db.Delete(elements...); err != nil {
		return err
	}
	return nil
}

// return a new User struct
func (u *UserRegistry) New(name, password, email string) *User {
	return &User{name, password, email, util.NewStringSet(), util.NewStringSet()}
}
