package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DAv10195/submit_commons/containers"
	submiterr "github.com/DAv10195/submit_commons/errors"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/assignments"
	"github.com/DAv10195/submit_server/elements/messages"
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
	if !exists {
		if _, err := NewUserBuilder(db.System, true).WithUserName(Admin).WithPassword(Admin).WithRoles(Admin).Build(); err != nil {
			return err
		}
	}
	return nil
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

// As building a user requires lots of validations and building of inner objects (e.g. a message box), the builder
// pattern can be really useful here
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
	withDbUpdate	 bool
}

// return a new User builder
func NewUserBuilder(asUser string, withDbUpdate bool) *UserBuilder{
	return &UserBuilder{roles: containers.NewStringSet(), coursesAsStaff: containers.NewStringSet(), coursesAsStudent: containers.NewStringSet(), asUser: asUser, withDbUpdate: withDbUpdate}
}

// set user name
func (b *UserBuilder) WithUserName(userName string) *UserBuilder {
	b.userName = userName
	return b
}

// set first name
func (b *UserBuilder) WithFirstName(firstName string) *UserBuilder {
	b.firstName = firstName
	return b
}

// set password
func (b *UserBuilder) WithPassword(password string) *UserBuilder {
	b.password = password
	return b
}

// set last name
func (b *UserBuilder) WithLastName(lastName string) *UserBuilder {
	b.lastName = lastName
	return b
}

// set email
func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.email = email
	return b
}

// add a course in which the built user will be a staff member
func (b *UserBuilder) WithCoursesAsStaff(CoursesAsStaff ...string) *UserBuilder {
	b.coursesAsStaff.Add(CoursesAsStaff...)
	return b
}

// add a course in which the built user will be a student
func (b *UserBuilder) WithCoursesAsStudent(CoursesAsStudent ...string) *UserBuilder {
	b.coursesAsStudent.Add(CoursesAsStudent...)
	return b
}

// add a role to the created user
func (b *UserBuilder) WithRoles(roles ...string)*UserBuilder{
	b.roles.Add(roles...)
	return b
}

// build the user, performing the required operations (db update, fs update) and validations
func (b *UserBuilder) Build() (*User, error) {
	if b.userName == "" {
		return nil, &submiterr.ErrInsufficientData{Message: "given user name can't be empty"}
	}
	exists, err := db.KeyExistsInBucket([]byte(db.Users), []byte(b.userName))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &db.ErrKeyExistsInBucket{Bucket: db.Users, Key: b.userName}
	}
	if b.password == "" {
		return nil, &submiterr.ErrInsufficientData{Message: "given password can't be empty"}
	}
	if b.roles.NumberOfElements() == 0 {
		return nil, &submiterr.ErrInsufficientData{Message: "user must have at least one role"}
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
	user := &User{
		UserName: b.userName,
		FirstName: b.firstName,
		LastName: b.lastName,
		Password: encryptedPassword,
		Email: b.email,
		Roles: b.roles,
		CoursesAsStaff: b.coursesAsStaff,
		CoursesAsStudent: b.coursesAsStudent,
	}
	if b.withDbUpdate {
		messageBox := messages.NewMessageBox()
		user.MessageBox = messageBox.ID
		if err := db.Update(b.asUser, messageBox, user); err != nil {
			return nil, err
		}
	}
	return user, nil
}

// delete the user also deleting his message box and assignment instances
func Delete(user *User, withFsUpdate bool) error {
	var instToDel []*assignments.AssignmentInstance
	if err := db.QueryBucket([]byte(db.AssignmentInstances), func(_, elemBytes []byte) error {
		inst := &assignments.AssignmentInstance{}
		if err := json.Unmarshal(elemBytes, inst); err != nil {
			return err
		}
		if inst.UserName == user.UserName {
			instToDel = append(instToDel, inst)
		}
		return nil
	}); err != nil {
		return err
	}
	for _, inst := range instToDel {
		if err := assignments.DeleteInstance(inst, withFsUpdate); err != nil {
			return err
		}
	}
	box, err := messages.Get(user.MessageBox)
	if err != nil {
		return err
	}
	if err := messages.Delete(box); err != nil {
		return err
	}
	return db.Delete(user)
}
