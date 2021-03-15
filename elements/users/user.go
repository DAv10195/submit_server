package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/messages"
	"github.com/DAv10195/submit_server/util"
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
		return &util.ErrInsufficientData{Message: "missing user name"}
	}
	exists, err := db.KeyExistsInBucket([]byte(db.Users), []byte(user.UserName))
	if err != nil {
		return err
	}
	if exists {
		return &db.ErrKeyExistsInBucket{Bucket: db.Users, Key: user.UserName}
	}
	if user.Password == "" {
		return &util.ErrInsufficientData{Message: "missing password"}
	}
	return nil
}

type UserBuilder struct {
	userName         string
	firstName        string
	lastName         string
	password         string
	email            string
	roles            *containers.StringSet
	coursesAsStaff   *containers.StringSet
	coursesAsStudent *containers.StringSet
	asUser			 string
}

func NewUserBuilder(asUser string) *UserBuilder{
	return &UserBuilder{roles: containers.NewStringSet(), coursesAsStaff: containers.NewStringSet(), coursesAsStudent: containers.NewStringSet(), asUser: asUser}
}

func (b *UserBuilder) WithUserName(userName string) *UserBuilder {
	b.userName = userName
	return b
}
func (b *UserBuilder) WithFirstName(firstName string) *UserBuilder {
	b.firstName = firstName
	return b
}
func (b *UserBuilder) WithPassword(password string) *UserBuilder {
	b.password = password
	return b
}
func (b *UserBuilder) WithLastName(lastName string) *UserBuilder {
	b.lastName = lastName
	return b
}
func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.email = email
	return b
}
func (b *UserBuilder) WithCoursesAsStaff(CoursesAsStaff ...string) *UserBuilder {
	b.coursesAsStaff.Add(CoursesAsStaff...)
	return b
}
func (b *UserBuilder) WithCoursesAsStudent(CoursesAsStudent ...string) *UserBuilder {
	b.coursesAsStudent.Add(CoursesAsStudent...)
	return b
}
func (b *UserBuilder) WithRoles(roles ...string)*UserBuilder{
	b.roles.Add(roles...)
	return b
}

func (b *UserBuilder) Build() (*User, error) {
	if b.userName == "" {
		return nil, &util.ErrInsufficientData{Message: "given user name can't be empty"}
	}
	exists, err := db.KeyExistsInBucket([]byte(db.Users), []byte(b.userName))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &db.ErrKeyExistsInBucket{Bucket: db.Users, Key: b.userName}
	}
	if b.password == "" {
		return nil, &util.ErrInsufficientData{Message: "given password can't be empty"}
	}
	if b.roles.NumberOfElements() == 0 {
		return nil, &util.ErrInsufficientData{Message: "user must have at least one role"}
	}
	for _, r := range b.roles.Slice() {
		if !Roles.Contains(r) {
			return nil, fmt.Errorf("invalid role: %s", r)
		}
	}
	if containers.StringSetIntersection(b.coursesAsStudent, b.coursesAsStaff).NumberOfElements() > 0 {
		return nil, errors.New("user can't be a staff member and a student in the same course")
	}
	allCoursesOfUser := containers.StringSetUnion(b.coursesAsStudent, b.coursesAsStaff)
	for _, course := range allCoursesOfUser.Slice() {
		exists, err := db.KeyExistsInBucket([]byte(db.Courses), []byte(course))
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, &db.ErrKeyNotFoundInBucket{Bucket: db.Courses, Key: course}
		}
	}
	encryptedPassword, err := db.Encrypt(b.password)
	if err != nil {
		return nil, err
	}
	messageBox := messages.NewMessageBox()
	user := &User{
		UserName: b.userName,
		FirstName: b.firstName,
		LastName: b.lastName,
		Password: encryptedPassword,
		Email: b.email,
		MessageBox: messageBox.ID,
		Roles: b.roles,
		CoursesAsStaff: b.coursesAsStaff,
		CoursesAsStudent: b.coursesAsStudent,
	}
	if err := db.Update(b.asUser, messageBox, user); err != nil {
		return nil, err
	}
	return user, nil
}
