package users

import (
	"github.com/DAv10195/submit_server/db"
	"os"
	"path/filepath"
	"testing"
)

func TestInitDefaultAdmin(t *testing.T) {
	path := os.TempDir()
	if err := db.InitDB(path); err != nil {
		t.Fatal(err)
	}
	defer func(){
		if err := os.Remove(filepath.Join(path, db.DatabaseFileName)); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(filepath.Join(path, db.DatabaseEncryptionKeyFileName)); err != nil {
			t.Fatal(err)
		}
	}()
	if err := InitDefaultAdmin(); err != nil {
		t.Fatal(err)
	}
	exists, err := db.KeyExistsInBucket([]byte(Users), []byte(Admin))
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatalf("\"%s\" user should exist in \"%s\" bucket but it doesn't", Admin, Users)
	}
}
