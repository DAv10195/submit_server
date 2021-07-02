package server

import "time"

const (
	ContentType 			= "Content-Type"
	ApplicationJson 		= "application/json"

	userName				= "userName"

	accessDenied			= "access denied"

	authenticatedUser		= "authenticated_user"

	agentId					= "agentId"
	hello					= "Hello"
	endpoint				= "endpoint"

	serverTimeout			= 15 * time.Second

	numTaskProcWorkers		= 10
	taskProcessingTimeout	= 120
	taskId					= "taskId"
	taskRespId				= "taskRespId"

	trueStr					= "true"

	courseNumber			= "courseNumber"
	courseYear				= "courseYear"

	assDefName				= "assDefName"

	testName				= "testName"
)
