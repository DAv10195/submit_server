package db

import (
	"github.com/boltdb/bolt"
	"os"
	"path/filepath"
)

var buckets = []string{Courses, Users}

var db *bolt.DB

func InitDB(path string) error {
	if err := initDbEncryption(filepath.Join(path, DatabaseEncryptionKeyFileName)); err != nil {
		return err
	}
	dbPath := filepath.Join(path, DatabaseFileName)
	logger.Infof("loading DB from %s ...", path)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		logger.Infof("DB doesn't exist in %s. Creating new one...", path)
	} else {
		logger.Error(err)
		return err
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
