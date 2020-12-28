package db

import (
	"encoding/json"
	"github.com/boltdb/bolt"
)

// update (or create, if they don't exist yet) the give elements in the DB
func Update(asUser string, elements ...IBucketElement) error {
	if len(elements) == 0 {
		return nil
	}
	return db.Update(func (tx *bolt.Tx) error {
		for _, element := range elements {
			bucket := element.Bucket()
			dbBucket := tx.Bucket(bucket)
			if dbBucket == nil {
				return &ErrBucketNotFound{string(bucket)}
			}
			key := element.Key()
			if dbBucket.Get(key) == nil {
				element.MarkInsert(asUser)
			} else {
				element.MarkUpdate(asUser)
			}
			objectBytes, err := json.Marshal(element)
			if err != nil {
				return err
			}
			if err := dbBucket.Put(key, objectBytes); err != nil {
				return err
			}
		}
		return nil
	})
}

// delete the given elements (if they exist) from the DB
func Delete(elements ...IBucketElement) error {
	if len(elements) == 0 {
		return nil
	}
	return db.Update(func (tx *bolt.Tx) error {
		for _, element := range elements {
			bucket := element.Bucket()
			dbBucket := tx.Bucket(bucket)
			if dbBucket == nil {
				return &ErrBucketNotFound{string(bucket)}
			}
			if err := dbBucket.Delete(element.Key()); err != nil {
				return err
			}
		}
		return nil
	})
}

// determines if the given key exists in the given bucket
func KeyExistsInBucket(bucket, key []byte) (bool, error) {
	exists := false
	err := db.View(func (tx *bolt.Tx) error {
		dbBucket := tx.Bucket(bucket)
		if dbBucket == nil {
			return &ErrBucketNotFound{string(bucket)}
		}
		exists = dbBucket.Get(key) != nil
		return nil
	})
	if err != nil {
		return false, err
	}
	return exists, nil
}