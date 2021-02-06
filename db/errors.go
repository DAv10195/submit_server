package db

import "fmt"

type ErrBucketNotFound struct {
	Bucket 	string
}

func (e *ErrBucketNotFound) Error() string {
	return fmt.Sprintf("%s bucket not found", e.Bucket)
}

type ErrKeyNotFoundInBucket struct {
	Bucket 	string
	Key 	string
}

func (e *ErrKeyNotFoundInBucket) Error() string {
	return fmt.Sprintf("%s key not found in bucket %s", e.Key, e.Bucket)
}

type ErrKeyExistsInBucket struct {
	Bucket 	string
	Key 	string
}

func (e *ErrKeyExistsInBucket) Error() string {
	return fmt.Sprintf("%s key already exists in bucket %s", e.Key, e.Bucket)
}
