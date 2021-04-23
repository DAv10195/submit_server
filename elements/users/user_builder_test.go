package users

import (
	"errors"
	"fmt"
	submiterr "github.com/DAv10195/submit_commons/errors"
	"github.com/DAv10195/submit_server/db"
	"testing"
)

func createValidUser() error {
	builder := NewUserBuilder(db.System, false)
	builder.WithEmail("nikita.kogan@sap.com").WithFirstName("nikita").
		WithLastName("kogan").WithUserName("nikita").WithPassword("nikita").
		WithRoles(Admin)
	_, err := builder.Build()
	if err != nil {
		return fmt.Errorf("error creating user with valid arguments: %v", err)
	}
	return nil
}

func emptyUserName() error {
	builder := NewUserBuilder(db.System, false)
	builder.WithEmail("nikita.kogan@sap.com").WithFirstName("nikita").
		WithLastName("kogan").WithPassword("nikita").
		WithRoles(Admin)
	_, err := builder.Build()
	if err == nil {
		return errors.New("error not returned when building user with empty user name")
	}
	if _, ok := err.(*submiterr.ErrInsufficientData); !ok {
		return errors.New("returned error is not of type ErrInsufficientData")
	}
	return nil
}

func emptyRoles() error {
	builder := NewUserBuilder(db.System, false)
	builder.WithEmail("nikita.kogan@sap.com").WithFirstName("nikita").
		WithLastName("kogan").WithUserName("nikita").WithPassword("nikita")
	_, err := builder.Build()
	if err == nil {
		return errors.New("error not returned when building user with no roles")
	}
	if _, ok := err.(*submiterr.ErrInsufficientData); !ok {
		return errors.New("returned error is not of type ErrInsufficientData")
	}
	return nil
}

func emptyPassword() error {
	builder := NewUserBuilder(db.System, false)
	builder.WithEmail("nikita.kogan@sap.com").WithFirstName("nikita").
		WithLastName("kogan").WithUserName("nikita").WithRoles(Admin)
	_, err := builder.Build()
	if err == nil {
		return errors.New("error not returned when building user with empty password")
	}
	if _, ok := err.(*submiterr.ErrInsufficientData); !ok {
		return errors.New("returned error is not of type ErrInsufficientData")
	}
	return nil
}

func TestBuilder(t *testing.T) {
	cleanup := db.InitDbForTest()
	defer cleanup()
	testFuncs := []func() error{emptyPassword, emptyRoles, emptyUserName, createValidUser}
	for _, testFunc := range testFuncs {
		if err := testFunc(); err != nil {
			t.Fatal(err)
		}
	}
}
