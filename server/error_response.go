package server

import "encoding/json"

// error response
type ErrorResponse struct {
	Message	string	`json:"message"`
}

func (e *ErrorResponse) String() string {
	errRespBytes, _ := json.Marshal(e)
	return string(errRespBytes)
}
