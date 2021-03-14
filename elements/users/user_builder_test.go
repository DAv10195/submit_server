package users

import (
	"errors"
	"github.com/DAv10195/submit_server/db"
	"testing"
)

func createValidUser(t *testing.T){
	builder := NewUserBuilder()
	builder.WithEmail("nikita.kogan@sap.com").WithFirstName("nikita").
		WithLastName("kogan").WithUserName("nikita").WithPassword("nikita").
		WithRoles(Admin).WithCoursesAsStaff("infi").WithCoursesAsStudent("algo")
	_,err := builder.Build()
	if err != nil {
		t.Fatal(errors.New("error create valid user test"))
	}
}

func emptyUserName(t *testing.T){
	builder := NewUserBuilder()
	builder.WithEmail("nikita.kogan@sap.com").WithFirstName("nikita").
		WithLastName("kogan").WithPassword("nikita").
		WithRoles(Admin).WithCoursesAsStaff("infi").WithCoursesAsStudent("algo")
	_,err := builder.Build()
	if err == nil {
		t.Fatal(errors.New("error emptyUserName test"))
	}
}

func emptyRoles(t *testing.T){
	builder := NewUserBuilder()
	builder.WithEmail("nikita.kogan@sap.com").WithFirstName("nikita").
		WithLastName("kogan").WithUserName("nikita").WithPassword("nikita").
		WithCoursesAsStaff("infi").WithCoursesAsStudent("algo")
	_,err := builder.Build()
	if err == nil {
		t.Fatal(errors.New("error emptyRoles test"))
	}
}

func emptyPassword(t *testing.T){
	builder := NewUserBuilder()
	builder.WithEmail("nikita.kogan@sap.com").WithFirstName("nikita").
		WithLastName("kogan").WithUserName("nikita").WithRoles(Admin).
		WithCoursesAsStaff("infi").WithCoursesAsStudent("algo")
	_,err := builder.Build()
	if err == nil {
		t.Fatal(errors.New("error emptyPassword test"))
	}
}

func TestBuilder(t *testing.T){
	_ = db.InitDbForTest()
	//defer cleanup()
	emptyPassword(t)
	emptyRoles(t)
	emptyUserName(t)
	createValidUser(t)
}
