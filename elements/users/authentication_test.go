package users

import (
	"github.com/DAv10195/submit_server/db"
	"testing"
)

func TestAuthenticate(t *testing.T) {
	cleanup := db.InitDbForTest()
	defer cleanup()
	if err := InitDefaultAdmin(); err != nil {
		t.Fatal(err)
	}
	if _, err := Authenticate(Admin, Admin); err != nil {
		t.Fatal(err)
	}
}
