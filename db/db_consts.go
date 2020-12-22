package db

import "time"

const (
	dbPerms = 0600
	dbOpenTimeout = time.Minute

	// roles
	Admin = "admin"

	// main buckets
	Users = "users"
	Courses = "courses"
)
