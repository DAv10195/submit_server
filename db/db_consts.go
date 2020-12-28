package db

import "time"

const (
	dbPerms = 0600
	dbOpenTimeout = time.Minute
	dbFileName = "submit_server.db"
	dbEncryptionKeyFileName = "submit_server.key"

	System = "system"

	// roles
	Admin = "admin"
	StandardUser = "std_user"

	// bucket names
	Users = "users"
	Courses = "courses"
	Assignments = "assignments"
	Submissions = "submissions"
	SubmissionResults = "submission_results"
)
