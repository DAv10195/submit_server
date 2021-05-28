// facade for accessing the DB
package db

import (
	"bytes"
	"encoding/json"
	"github.com/boltdb/bolt"
)

// update (or create, if they don't exist yet) the given elements in the DB
func Update(asUser string, elements ...IBucketElement) error {
	if len(elements) == 0 {
		return nil
	}
	return db.Update(func (tx *bolt.Tx) error {
		for _, element := range elements {
			bucket := element.Bucket()
			dbBucket := tx.Bucket(bucket)
			if dbBucket == nil {
				err := &ErrBucketNotFound{string(bucket)}
				logger.WithError(err).Errorf("error updating \"%s\" bucket", string(bucket))
				return err
			}
			key := element.Key()
			if dbBucket.Get(key) == nil {
				logger.Debugf("inserting element with key = \"%s\" into \"%s\" bucket", string(key), string(bucket))
				element.MarkInsert(asUser)
			} else {
				logger.Debugf("updating element with key = \"%s\" in \"%s\" bucket", string(key), string(bucket))
				element.MarkUpdate(asUser)
			}
			objectBytes, err := json.Marshal(element)
			if err != nil {
				logger.WithError(err).Errorf("error updating key = \"%s\" in \"%s\" bucket", string(key), string(bucket))
				return err
			}
			if err := dbBucket.Put(key, objectBytes); err != nil {
				logger.WithError(err).Errorf("error updating key = \"%s\" in \"%s\" bucket", string(key), string(bucket))
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
				err := &ErrBucketNotFound{string(bucket)}
				logger.WithError(err).Errorf("error deleting elements from \"%s\" bucket", string(bucket))
				return err
			}
			key := element.Key()
			if err := dbBucket.Delete(key); err != nil {
				logger.WithError(err).Errorf("error deleting key = \"%s\" from \"%s\" bucket", string(key), string(bucket))
				return err
			}
		}
		return nil
	})
}

// delete the given keys from the given bucket
func DeleteKeysFromBucket(bucket []byte, keys ...[]byte) error {
	if len(keys) == 0 {
		return nil
	}
	return db.Update(func (tx *bolt.Tx) error {
		dbBucket := tx.Bucket(bucket)
		if dbBucket == nil {
			err := &ErrBucketNotFound{string(bucket)}
			logger.WithError(err).Errorf("error deleting keys from \"%s\" bucket", string(bucket))
			return err
		}
		for _, key := range keys {
			if err := dbBucket.Delete(key); err != nil {
				logger.WithError(err).Errorf("error deleting key = \"%s\" from \"%s\" bucket", string(key), string(bucket))
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
			err := &ErrBucketNotFound{string(bucket)}
			logger.WithError(err).Errorf("error querying \"%s\" bucket", string(bucket))
			return err
		}
		exists = dbBucket.Get(key) != nil
		return nil
	})
	if err != nil {
		return false, err
	}
	return exists, nil
}

// a function that accepts the bucket element key and data and processes it using the implemented strategy
type BucketElementProcessingFunc func([]byte, []byte) error

// given a bucket and a processing function, process all elements in that bucket
func QueryBucket(bucket []byte, process BucketElementProcessingFunc) error {
	return db.View(func (tx *bolt.Tx) error {
		dbBucket := tx.Bucket(bucket)
		if dbBucket == nil {
			err := &ErrBucketNotFound{string(bucket)}
			logger.WithError(err).Errorf("error querying \"%s\" bucket", string(bucket))
			return err
		}
		dbCursor := dbBucket.Cursor()
		for elementKey, elementBytes := dbCursor.First(); elementKey != nil; elementKey, elementBytes = dbCursor.Next() {
			if err := process(elementKey, elementBytes); err != nil {
				if _, ok := err.(*ErrStopQuery); ok {
					elementKey, elementBytes = dbCursor.Next()
					if elementKey != nil {
						return &ErrElementsLeftToProcess{}
					}
					return nil
				}
				logger.WithError(err).Errorf("error querying \"%s\" bucket", string(bucket))
				return err
			}
		}
		return nil
	})
}

// given a bucket and a key, return the bytes of the data assigned with that key
func GetFromBucket(bucket, key []byte) ([]byte, error) {
	var data bytes.Buffer
	if err := db.View(func (tx *bolt.Tx) error {
		dbBucket := tx.Bucket(bucket)
		if dbBucket == nil {
			err := &ErrBucketNotFound{string(bucket)}
			logger.WithError(err).Errorf("error accessing \"%s\" bucket", string(bucket))
			return err
		}
		bytesOfKey := dbBucket.Get(key)
		if bytesOfKey == nil {
			err := &ErrKeyNotFoundInBucket{string(bucket), string(key)}
			logger.WithError(err).Errorf("error accessing \"%s\" key in \"%s\" bucket", string(key), string(bucket))
			return err
		}
		if _, err := data.Write(bytesOfKey); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}
