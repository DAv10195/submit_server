package users

import (
	"encoding/json"
	"fmt"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/messages"
	"github.com/DAv10195/submit_server/util/containers"
)

// user
type User struct {
	db.ABucketElement
	UserName       			string                	`json:"user_name"`
	FirstName				string					`json:"first_name"`
	LastName				string					`json:"last_name"`
	Password   				string                	`json:"password"`
	Email      				string                	`json:"email"`
	MessageBox 				string                	`json:"message_box"`
	Roles      				*containers.StringSet 	`json:"roles"`
	CoursesAsStaff			*containers.StringSet 	`json:"courses_as_staff"`
	CoursesAsStudent		*containers.StringSet	`json:"courses_as_student"`
}

func (u *User) Key() []byte {
	return []byte(u.UserName)
}

func (u *User) Bucket() []byte {
	return []byte(db.Users)
}

// check if the default admin user is present in the DB and add it if not
func InitDefaultAdmin() error {
	exists, err := db.KeyExistsInBucket([]byte(db.Users), []byte(Admin))
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
	messageBox := messages.NewMessageBox()
	user := &User{
		UserName: Admin,
		Password: password,
		MessageBox: messageBox.ID,
		Roles: containers.NewStringSet(),
		CoursesAsStaff: containers.NewStringSet(),
		CoursesAsStudent: containers.NewStringSet(),
	}
	user.Roles.Add(Admin)
	return db.Update(db.System, messageBox, user)
}

// return the user represented by the given user name if that user exists
func Get(userName string) (*User, error) {
	userBytes, err := db.GetFromBucket([]byte(db.Users), []byte(userName))
	if err != nil {
		return nil, err
	}
	user := &User{}
	if err = json.Unmarshal(userBytes, user); err != nil {
		return nil, err
	}

	return user, nil
}

// authenticate the user with the given password. Returns the authenticated user when the returned error is nil
func Authenticate(user, password string) (*User, error) {
	userStruct, err := Get(user)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			return nil, &ErrAuthenticationFailure{user, fmt.Sprintf("user \"%s\" not found", user)}
		}
		return nil, err
	}
	userPassword, err := db.Decrypt(userStruct.Password)
	if err != nil {
		return nil, err
	}
	if password != userPassword {
		return nil, &ErrAuthenticationFailure{user, "incorrect password"}
	}
	return userStruct, nil
}

func ValidateNew(user *User) error {
	if user.UserName == "" {
		return &ErrInsufficientData{"missing user name"}
	}
	exists, err := db.KeyExistsInBucket([]byte(db.Users), []byte(user.UserName))
	if err != nil {
		return err
	}
	if exists {
		return &db.ErrKeyExistsInBucket{Bucket: db.Users, Key: user.UserName}
	}
	if user.Password == "" {
		return &ErrInsufficientData{"missing password"}
	}
	return nil
}
