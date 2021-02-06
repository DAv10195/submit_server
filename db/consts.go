package db

import "time"

const (
	dbPerms                       = 0600
	dbOpenTimeout                 = time.Minute
	DatabaseFileName              = "submit_server.db"
	DatabaseEncryptionKeyFileName = "submit_server.key"

	System = "system"

	KeySeparator				= ":"

	// bucket names
	Users	 					= "users"
	Courses 					= "courses"
	AssignmentInstances 		= "assignment_instances"
	AssignmentDefinitions		= "assignment_definitions"
	MessageBoxes				= "message_boxes"
	Messages					= "messages"
	Tests						= "tests"
	Appeals						= "appeals"
)
