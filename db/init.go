package db

import (
	"github.com/boltdb/bolt"
	"os"
	"path/filepath"
)

var buckets = []string{Courses, Users, Assignments, Submissions, SubmissionResults}

var db *bolt.DB

func initDB(path string) error {
	if err := initDbEncryption(filepath.Join(path, dbEncryptionKeyFileName)); err != nil {
		return err
	}
	dbPath := filepath.Join(path, dbFileName)
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
	if err := initAdminUser(); err != nil {
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

func initAdminUser() error {
	adminUser, err := GetDefaultAdmin()
	if err != nil {
		return err
	}
	logger.Debugf("validating existence of user \"%s\"", adminUser.Name)
	exists, err := KeyExistsInBucket(adminUser.Bucket(), adminUser.Key())
	if err != nil {
		return err
	}
	if !exists {
		return Update(System, adminUser)
	}
	return nil
}
