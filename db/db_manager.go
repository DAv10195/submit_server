package db

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"os"
)

// manages all operations against the Bolt DB
type BoltDBManager struct {
	path	string
	boltDb	*bolt.DB
}

// initialize the Bolt DB such that it contains all of the basic buckets required
func (db *BoltDBManager) init() error {
	if err := db.initUsers(); err != nil {
		return err
	}
	if err := db.initCourses(); err != nil {
		return err
	}
	return nil
}

// initialize the users bucket and the basic administrator user
func (db *BoltDBManager) initUsers() error {
	return db.boltDb.Update(func (tx *bolt.Tx) error {
		users, err := tx.CreateBucketIfNotExists([]byte(Users))
		if err != nil {
			return err
		}
		adminKey := []byte(Admin)
		if users.Get(adminKey) == nil {
			defAdminUser, err := json.Marshal(getDefaultAdmin())
			if err != nil {
				return err
			}
			if err := users.Put(adminKey, defAdminUser); err != nil {
				return err
			}
		}
		return nil
	})
}

// initialize the courses bucket
func (db *BoltDBManager) initCourses() error {
	return db.boltDb.Update(func (tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(Courses))
		if err != nil {
			return err
		}
		return nil
	})
}

// load the DB from the given path
func (db *BoltDBManager) Load() error {
	logger.Infof("Loading DB from %s...", db.path)
	if _, err := os.Stat(db.path); os.IsNotExist(err) {
		logger.Infof("DB doesn't exist in %s. Creating new one...", db.path)
	} else {
		logger.Fatalf("Error accessing existing DB in %s", db.path)
	}
	boltDb, err := bolt.Open(db.path, dbPerms, &bolt.Options{Timeout: dbOpenTimeout})
	if err != nil {
		logger.WithError(err).Errorf("failed to load DB from %s", db.path)
		return err
	}
	db.boltDb = boltDb
	if err := db.init(); err != nil {
		logger.WithError(err).Errorf("failed to initialize DB at %s", db.path)
		return err
	}
	logger.Infof("DB loaded successfully from %s", db.path)
	return nil
}

// fills the given elements with the data stored for them in the DB
func (db *BoltDBManager) Get(elements ...BucketElement) error {
	if len(elements) == 0 {
		return nil
	}
	return db.boltDb.View(func (tx *bolt.Tx) error {
		for _, element := range elements {
			bucket := element.Bucket()
			dbBucket := tx.Bucket(bucket)
			if dbBucket == nil {
				return &ErrBucketNotFound{string(bucket)}
			}
			key := element.Key()
			objectBytes := dbBucket.Get(key)
			if objectBytes == nil {
				return &ErrKeyNotFoundInBucket{string(key), string(bucket)}
			}
			if err := json.Unmarshal(objectBytes, element); err != nil {
				return err
			}
		}
		return nil
	})
}

// updates/creates the given elements in the DB
func (db *BoltDBManager) Put(elements ...BucketElement) error {
	if len(elements) == 0 {
		return nil
	}
	return db.boltDb.Update(func (tx *bolt.Tx) error {
		for _, element := range elements {
			bucket := element.Bucket()
			dbBucket := tx.Bucket(bucket)
			if dbBucket == nil {
				return &ErrBucketNotFound{string(bucket)}
			}
			objectBytes, err := json.Marshal(element)
			if err != nil {
				return err
			}
			if err := dbBucket.Put(element.Key(), objectBytes); err != nil {
				return err
			}
		}
		return nil
	})
}

// delete the given elements from the DB if they are present in it
func (db *BoltDBManager) Delete(elements ...BucketElement) error {
	if len(elements) == 0 {
		return nil
	}
	return db.boltDb.Update(func (tx *bolt.Tx) error {
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
