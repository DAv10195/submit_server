package users

import (
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
		t.Fatal(err)
	}
}

func emptyUserName(t *testing.T){
	builder := NewUserBuilder()
	builder.WithEmail("nikita.kogan@sap.com").WithFirstName("nikita").
		WithLastName("kogan").WithPassword("nikita").
		WithRoles(Admin).WithCoursesAsStaff("infi").WithCoursesAsStudent("algo")
	_,err := builder.Build()
	if err == nil {
		t.Fatal(err)
	}
}

func emptyRoles(t *testing.T){
	builder := NewUserBuilder()
	builder.WithEmail("nikita.kogan@sap.com").WithFirstName("nikita").
		WithLastName("kogan").WithUserName("nikita").WithPassword("nikita").
		WithCoursesAsStaff("infi").WithCoursesAsStudent("algo")
	_,err := builder.Build()
	if err == nil {
		t.Fatal(err)
	}
}

func emptyPassword(t *testing.T){
	builder := NewUserBuilder()
	builder.WithEmail("nikita.kogan@sap.com").WithFirstName("nikita").
		WithLastName("kogan").WithUserName("nikita").WithRoles(Admin).
		WithCoursesAsStaff("infi").WithCoursesAsStudent("algo")
	_,err := builder.Build()
	if err == nil {
		t.Fatal(err)
	}
}

func TestBuilder(t *testing.T){
	cleanup := db.InitDbForTest()
	defer cleanup()
	emptyPassword(t)
	emptyRoles(t)
	emptyUserName(t)
	createValidUser(t)
}
