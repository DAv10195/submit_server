package users

import "fmt"

type ErrAuthenticationFailure struct {
	User	string
	Message	string
}

func (e *ErrAuthenticationFailure) Error() string {
	return fmt.Sprintf("error authenticating user \"%s\": %s", e.User, e.Message)
}

type ErrEmailValidationFailed struct {
	Email	string
	Status	string
}

func (e *ErrEmailValidationFailed) Error() string {
	return fmt.Sprintf("validation of \"%s\" email failed with status \"%s\"", e.Email, e.Status)
}
