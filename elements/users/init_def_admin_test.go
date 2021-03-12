package users

import (
	"github.com/DAv10195/submit_server/db"
	"testing"
)

func TestInitDefaultAdmin(t *testing.T) {
	cleanup := db.InitDbForTest()
	defer cleanup()
	if err := InitDefaultAdmin(); err != nil {
		t.Fatal(err)
	}
	exists, err := db.KeyExistsInBucket([]byte(db.Users), []byte(Admin))
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatalf("\"%s\" user should exist in \"%s\" bucket but it doesn't", Admin, db.Users)
	}
}
