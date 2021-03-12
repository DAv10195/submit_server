package db

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	path := filepath.Join(os.TempDir(), DatabaseFileName)
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
	mockElement1, mockElement2 := &mockBucketElement{Field: fmt.Sprintf("%s1", mock)}, &mockBucketElement{Field: fmt.Sprintf("%s2", mock)}
	if err := Update(System, mockElement1, mockElement2); err != nil {
		t.Fatal(err)
	}
	exists, err := KeyExistsInBucket(mockElement1.Bucket(), mockElement1.Key())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("mock element 1 wasn't found in the DB")
	}
	exists, err = KeyExistsInBucket(mockElement2.Bucket(), mockElement2.Key())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("mock element 2 wasn't found in the DB")
	}
	if err := Delete(mockElement1); err != nil {
		t.Fatal(err)
	}
	exists, err = KeyExistsInBucket(mockElement1.Bucket(), mockElement1.Key())
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("mock element1 is still in the DB after deleting using Delete")
	}
	if err := DeleteKeysFromBucket(mockElement2.Bucket(), mockElement2.Key()); err != nil {
		t.Fatal(err)
	}
	exists, err = KeyExistsInBucket(mockElement2.Bucket(), mockElement2.Key())
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("mock element2 is still in the DB after deleting using DeleteKeysFromBucket")
	}
}

func TestQueryAndGet(t *testing.T) {
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
	mockElementFromDb1, mockElementFromDb2 := &mockBucketElement{}, &mockBucketElement{}
	if err := QueryBucket(mockElement.Bucket(), func (key, data []byte) error {
		if bytes.Compare(key, mockElement.Key()) == 0 {
			if err := json.Unmarshal(data, mockElementFromDb1); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if mockElement.Field != mockElementFromDb1.Field {
		t.Fatalf("expected the query to return the same element but it didn't")
	}
	mockElementFromDb2Bytes, err := GetFromBucket(mockElement.Bucket(), mockElement.Key())
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(mockElementFromDb2Bytes, mockElementFromDb2); err != nil {
		t.Fatal(err)
	}
	if mockElement.Field != mockElementFromDb2.Field {
		t.Fatalf("expected get to return the same element but it didn't")
	}
}
