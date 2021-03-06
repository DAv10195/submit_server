package server

import "time"

const (
	ContentType 			= "Content-Type"
	ApplicationJson 		= "application/json"

	userName				= "userName"

	accessDenied			= "access denied"

	authenticatedUser		= "authenticated_user"

	serverTimeout			= 15 * time.Second
)
