package db

import (
	"fmt"
	commons "github.com/DAv10195/submit_commons"
	"github.com/boltdb/bolt"
	"os"
	"path/filepath"
)

var buckets = []string{Courses, Users, AssignmentInstances, AssignmentDefinitions, MessageBoxes, Messages, Tests, Appeals, Agents, Tasks, TaskResponses}

var db *bolt.DB

// initialize the BoltDB in the given path. In case the DB exists in the given path, then the existing DB will be used
func InitDB(path string) error {
	if err := initDbEncryption(filepath.Join(path, DatabaseEncryptionKeyFileName)); err != nil {
		return err
	}
	dbPath := filepath.Join(path, DatabaseFileName)
	logger.Infof("loading DB from %s ...", path)
	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			logger.Infof("DB doesn't exist in %s. Creating new one...", path)
		} else {
			logger.WithError(err).Errorf("failed to initialize DB at %s", path)
			return err
		}
	}
	boltDb, err := bolt.Open(dbPath, dbPerms, &bolt.Options{Timeout: dbOpenTimeout})
	if err != nil {
		logger.WithError(err).Errorf("failed to load DB from %s", path)
		return err
	}
	db = boltDb
	if err := initBuckets(); err != nil {
		logger.WithError(err).Errorf("failed to initialize DB at %s", path)
		return err
	}
	logger.Infof("DB loaded successfully from %s", path)
	return nil
}

func initBuckets() error {
	return db.Update(func (tx *bolt.Tx) error {
		for _, bucket := range buckets {
			logger.Debugf("validating existence of bucket \"%s\"", bucket)
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// initializes a DB for testing and returns a cleanup function
func InitDbForTest() func() {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("submit_test_db_%s", commons.GenerateUniqueId()))
	if err := os.MkdirAll(path, 0755); err != nil {
		panic(err)
	}
	if err := InitDB(path); err != nil {
		panic(err)
	}
	return func() {
		if err := os.RemoveAll(path); err != nil {
			panic(err)
		}
	}
}
