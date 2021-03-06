package users

import "fmt"

type ErrAuthenticationFailure struct {
	User	string
	Message	string
}

func (e *ErrAuthenticationFailure) Error() string {
	return fmt.Sprintf("error authenticating user \"%s\": %s", e.User, e.Message)
}

type ErrInsufficientData struct {
	Message	string
}

func (e *ErrInsufficientData) Error() string {
	return e.Message
}
