package db

import "github.com/DAv10195/submit_server/util"

// user struct
type User struct {
	Name		string			`json:"name"`
	Password	string			`json:"password"`
	Email		string			`json:"email"`
	Courses		*util.StringSet	`json:"courses"`
	Roles		*util.StringSet	`json:"roles"`
}

func (u *User) Key() []byte {
	return []byte(u.Name)
}

func (u *User) Bucket() []byte {
	return []byte(Users)
}

// return a default admin user for DB initialization purposes. This is the default user in the system which has the
// admin role
func getDefaultAdmin() *User {
	user := &User{Admin, Admin, "", util.NewStringSet(), util.NewStringSet()}
	user.Roles.Add(Admin)
	return user
}
