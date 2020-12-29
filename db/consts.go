package db

import "time"

const (
	dbPerms                       = 0600
	dbOpenTimeout                 = time.Minute
	DatabaseFileName              = "submit_server.db"
	DatabaseEncryptionKeyFileName = "submit_server.key"

	System = "system"

	// bucket names
	Users	 	= "users"
	Courses 	= "courses"
)
