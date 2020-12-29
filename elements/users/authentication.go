package users

import (
	"encoding/json"
	"fmt"
	"github.com/DAv10195/submit_server/db"
)

// authenticate the user with the given password
func Authenticate(user, password string) error {
	userBytes, err := db.GetFromBucket([]byte(db.Users), []byte(user))
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			return &ErrAuthenticationFailure{user, fmt.Sprintf("user \"%s\" not found", user)}
		}
		return err
	}
	userStruct := &User{}
	if err := json.Unmarshal(userBytes, userStruct); err != nil {
		return err
	}
	userPassword, err := db.Decrypt(userStruct.Password)
	if err != nil {
		return err
	}
	if password != userPassword {
		return &ErrAuthenticationFailure{user, "incorrect password"}
	}
	return nil
}
