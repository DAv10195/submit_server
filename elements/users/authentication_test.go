package users

import (
	"github.com/DAv10195/submit_server/db"
	"os"
	"path/filepath"
	"testing"
)

func TestAuthenticate(t *testing.T) {
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
	if err := Authenticate(Admin, Admin); err != nil {
		t.Fatal(err)
	}
}
