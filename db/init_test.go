package db

import (
	"fmt"
	"github.com/boltdb/bolt"
	"os"
	"path/filepath"
	"testing"
)

func verifyBuckets() error {
	return db.View(func (tx *bolt.Tx) error {
		for _, bucket := range buckets {
			dbBucket := tx.Bucket([]byte(bucket))
			if dbBucket == nil {
				return fmt.Errorf("\"%s\" bucket should exist, but it doesn't", bucket)
			}
		}
		return nil
	})
}

func TestInit(t *testing.T) {
	path := os.TempDir()
	if err := InitDB(path); err != nil {
		t.Fatal(err)
	}
	defer func(){
		if err := os.Remove(filepath.Join(path, DatabaseFileName)); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(filepath.Join(path, DatabaseEncryptionKeyFileName)); err != nil {
			t.Fatal(err)
		}
	}()
	if err := verifyBuckets(); err != nil {
		t.Fatal(err)
	}
}
