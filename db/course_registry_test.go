package db

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCourseUpdateAndGet(t *testing.T) {
	dbManager := &BoltDBManager{filepath.Join(os.TempDir(), "test.db"), nil}
	if err := dbManager.Load(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(dbManager.path); err != nil {
			t.Fatal(err)
		}
	}()
	courseRegistry := &CourseRegistry{dbManager}
	course := courseRegistry.New("11111", "Course")
	if err := courseRegistry.Update(course); err != nil {
		t.Fatal(err)
	}
	coursesFromDB, err := courseRegistry.Get("11111")
	if err != nil {
		t.Fatal(err)
	}
	courseFromDB := coursesFromDB[0]
	if !reflect.DeepEqual(course, courseFromDB) {
		t.Fatalf("got\n\n%+v\n\nexpected\n\n%+v", courseFromDB, course)
	}
}

func TestCourseDelete(t *testing.T) {
	dbManager := &BoltDBManager{filepath.Join(os.TempDir(), "test.db"), nil}
	if err := dbManager.Load(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(dbManager.path); err != nil {
			t.Fatal(err)
		}
	}()
	courseRegistry := &CourseRegistry{dbManager}
	course := courseRegistry.New("11111", "Course")
	if err := courseRegistry.Update(course); err != nil {
		t.Fatal(err)
	}
	if err := courseRegistry.Delete(course.ID); err != nil {
		t.Fatal(err)
	}
	_, err := courseRegistry.Get(course.ID)
	if err == nil {
		t.Fatal("expected error be returned but got nil instead")
	}
	if _, ok := err.(*ErrKeyNotFoundInBucket); !ok {
		t.Fatal("err is not of type ErrKeyNotFoundInBucket")
	}
}
