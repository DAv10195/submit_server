package users

import (
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/util/containers"
)

// user struct
type User struct {
	db.ABucketElement
	Name		string                  `json:"name"`
	Password	string                  `json:"password"`
	Email		string                  `json:"email"`
	Courses		*containers.StringSet 	`json:"courses"`
	Roles		*containers.StringSet   `json:"roles"`
}

func (u *User) Key() []byte {
	return []byte(u.Name)
}

func (u *User) Bucket() []byte {
	return []byte(Users)
}

// check if the default admin user is present in the DB and add it if not
func InitDefaultAdmin() error {
	exists, err := db.KeyExistsInBucket([]byte(Users), []byte(Admin))
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	password, err := db.Encrypt(Admin)
	if err != nil {
		return err
	}
	user := &User{
		Name:     Admin,
		Password: password,
		Courses:  containers.NewStringSet(),
		Roles:    containers.NewStringSet(),
	}
	user.Roles.Add(Admin, StandardUser)
	return db.Update(db.System, user)
}
