package db

import "time"

const (
	dbPerms = 0600
	dbOpenTimeout = time.Minute

	// roles
	Admin = "admin"

	// bucket names
	Users = "users"
	Courses = "courses"
)
