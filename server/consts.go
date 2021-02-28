package server

import "time"

const (
	ContentType 			= "Content-Type"
	ApplicationJson 		= "application/json"
	Authorization			= "Authorization"

	userName				= "userName"
	logHttpErrFormat		= "error serving http request for %s"
	accessDenied			= "access denied"
	register				= "register"

	serverTimeout			= 15 * time.Second
)
