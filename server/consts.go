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

	trueStr					= "true"

	courseNumber			= "courseNumber"
	courseYear				= "courseYear"

	assDefName				= "assDefName"

	testName				= "testName"

	onDemandTask			= "on_demand_task"
	testTask				= "test_task"
	assInstUsrName			= "ass_inst_user_name"
	onSubmitExec			= "on_submit_exec"

	mossCopyThreshold		= "moss_copy_threshold"
)
