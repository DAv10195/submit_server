package db

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestUserUpdateAndGet(t *testing.T) {
	dbManager := &BoltDBManager{filepath.Join(os.TempDir(), "test.db"), nil}
	if err := dbManager.Load(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(dbManager.path); err != nil {
			t.Fatal(err)
		}
	}()
	userRegistry := &UserRegistry{dbManager}
	user := userRegistry.New("user", "user", "user@submit.com")
	if err := userRegistry.Update(user); err != nil {
		t.Fatal(err)
	}
	usersFromDB, err := userRegistry.Get("user")
	if err != nil {
		t.Fatal(err)
	}
	userFromDB := usersFromDB[0]
	if !reflect.DeepEqual(user, userFromDB) {
		t.Fatalf("got\n\n%+v\n\nexpected\n\n%+v", userFromDB, user)
	}
}

func TestUserDelete(t *testing.T) {
	dbManager := &BoltDBManager{filepath.Join(os.TempDir(), "test.db"), nil}
	if err := dbManager.Load(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(dbManager.path); err != nil {
			t.Fatal(err)
		}
	}()
	userRegistry := &UserRegistry{dbManager}
	user := userRegistry.New("user", "user", "user@submit.com")
	if err := userRegistry.Update(user); err != nil {
		t.Fatal(err)
	}
	if err := userRegistry.Delete(user.Name); err != nil {
		t.Fatal(err)
	}
	_, err := userRegistry.Get(user.Name)
	if err == nil {
		t.Fatal("expected error be returned but got nil instead")
	}
	if _, ok := err.(*ErrKeyNotFoundInBucket); !ok {
		t.Fatal("err is not of type ErrKeyNotFoundInBucket")
	}
}
