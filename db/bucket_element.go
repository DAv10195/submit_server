package db

// implementors of this interface specify to which Bucket they should be put and what should be their unique Key
type BucketElement interface {
	// get the key that should be associated with this element
	Key() []byte
	// get the name of the bucket that this element should be stored in
	Bucket() []byte
}
