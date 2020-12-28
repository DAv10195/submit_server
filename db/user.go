package db

import "github.com/DAv10195/submit_server/util/containers"

// user struct
type User struct {
	ABucketElement
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

// return a default admin user for DB initialization purposes. This is the default user in the system which has the
// admin role
func GetDefaultAdmin() (*User, error) {
	password, err := Encrypt(Admin)
	if err != nil {
		return nil, err
	}
	user := &User{
		Name:     Admin,
		Password: password,
		Courses:  containers.NewStringSet(),
		Roles:    containers.NewStringSet(),
	}
	user.Roles.Add(Admin, StandardUser)
	return user, nil
}
