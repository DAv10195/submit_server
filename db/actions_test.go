package db

import (
	"github.com/boltdb/bolt"
	"os"
	"path/filepath"
	"testing"
)


const mock = "mock"

type mockBucketElement struct {
	ABucketElement
	Field	string	`json:"field"`
}

func (m *mockBucketElement) Bucket() []byte {
	return []byte(mock)
}

func (m *mockBucketElement) Key() []byte {
	return []byte(m.Field)
}

func setDbWithMockBucket() (string, error) {
	path := filepath.Join(os.TempDir(), dbFileName)
	testDB, err := bolt.Open(path, dbPerms, &bolt.Options{Timeout: dbOpenTimeout})
	if err != nil {
		return "", err
	}
	db = testDB
	if err := db.Update(func (tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(mock)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return "", err
	}
	return path, nil
}

func TestUpdateAndDelete(t *testing.T) {
	dbPath, err := setDbWithMockBucket()
	if err != nil {
		t.Fatal(err)
	}
	defer func(){
		if err := os.Remove(dbPath); err != nil {
			t.Fatal(err)
		}
	}()
	mockElement := &mockBucketElement{Field: mock}
	if err := Update(System, mockElement); err != nil {
		t.Fatal(err)
	}
	exists, err := KeyExistsInBucket(mockElement.Bucket(), mockElement.Key())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("mock element wasn't found in the DB")
	}
	if err := Delete(mockElement); err != nil {
		t.Fatal(err)
	}
	exists, err = KeyExistsInBucket(mockElement.Bucket(), mockElement.Key())
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("mock element is still in the DB")
	}
}
