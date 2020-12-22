package db

// interface
type BucketElement interface {
	// get the key that should be associated with this element
	Key() []byte
	// get the name of the bucket that this element should be stored in
	Bucket() []byte
}
