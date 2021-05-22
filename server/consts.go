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
	taskProcessingTimeout	= 2 * time.Minute
	taskId					= "taskId"
	taskRespId				= "taskRespId"
)
