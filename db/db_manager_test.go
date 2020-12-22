package db

import (
	"github.com/DAv10195/submit_server/util"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	dbManager := &BoltDBManager{filepath.Join(os.TempDir(), "test.db"), nil}
	defer func() {
		if err := os.Remove(dbManager.path); err != nil {
			t.Fatal(err)
		}
	}()
	if err := dbManager.Load(); err != nil {
		t.Fatal(err)
	}
	adminFromDB := &User{Admin, "", "", nil, nil}
	if err := dbManager.Get(adminFromDB); err != nil {
		t.Fatal(err)
	}
	defaultAdmin := getDefaultAdmin()
	if !reflect.DeepEqual(adminFromDB, defaultAdmin) {
		t.Fatalf("got\n\n%+v\n\nexpected\n\n%+v", adminFromDB, defaultAdmin)
	}
}

func TestPutAndGet(t *testing.T) {
	dbManager := &BoltDBManager{filepath.Join(os.TempDir(), "test.db"), nil}
	defer func() {
		if err := os.Remove(dbManager.path); err != nil {
			t.Fatal(err)
		}
	}()
	if err := dbManager.Load(); err != nil {
		t.Fatal(err)
	}
	course := &Course{"89385", "CS Projects", util.NewStringSet(), util.NewStringSet()}
	student1 := &User{"david", "david", "david@submit.com", util.NewStringSet(), util.NewStringSet()}
	student2 := &User{"nikita", "nikita", "nikita@submit.com", util.NewStringSet(), util.NewStringSet()}
	instructor := &User{"osnat", "osnat", "osnat@submit.com", util.NewStringSet(), util.NewStringSet()}
	student1.Courses.Add(course.ID)
	student2.Courses.Add(course.ID)
	instructor.Courses.Add(course.ID)
	course.Students.Add(student1.Name, student2.Name)
	course.Staff.Add(instructor.Name)
	if err := dbManager.Put(course, student1, student2, instructor); err != nil {
		t.Fatal(err)
	}
	courseFromDB := &Course{"89385", "", nil, nil}
	student1FromDB := &User{"david", "", "", nil, nil}
	student2FromDB := &User{"nikita","","", nil, nil}
	instructorFromDB := &User{"osnat", "","", nil, nil}
	if err := dbManager.Get(courseFromDB, student1FromDB, student2FromDB, instructorFromDB); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(course, courseFromDB) {
		t.Fatalf("got\n\n%+v\n\nexpected\n\n%+v", course, courseFromDB)
	}
	if !reflect.DeepEqual(student1, student1FromDB) {
		t.Fatalf("got\n\n%+v\n\nexpected\n\n%+v", student1, student1FromDB)
	}
	if !reflect.DeepEqual(student2, student2FromDB) {
		t.Fatalf("got\n\n%+v\n\nexpected\n\n%+v", student2, student2FromDB)
	}
	if !reflect.DeepEqual(instructor, instructorFromDB) {
		t.Fatalf("got\n\n%+v\n\nexpected\n\n%+v", instructor, instructorFromDB)
	}
}

func TestDelete(t *testing.T) {
	dbManager := &BoltDBManager{filepath.Join(os.TempDir(), "test.db"), nil}
	defer func() {
		if err := os.Remove(dbManager.path); err != nil {
			t.Fatal(err)
		}
	}()
	if err := dbManager.Load(); err != nil {
		t.Fatal(err)
	}
	adminFromDB := &User{Admin, "", "", nil, nil}
	if err := dbManager.Delete(adminFromDB); err != nil {

	}
	err := dbManager.Get(adminFromDB)
	if err == nil {
		t.Fatal("expected an error but didn't get one when getting admin user from DB")
	}
	if _, ok := err.(*ErrKeyNotFoundInBucket); !ok {
		t.Fatal("expected error to be of type ErrKeyNotFoundInBucket but it is not")
	}
}