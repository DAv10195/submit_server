package server

import (
	"encoding/json"
	"fmt"
	"github.com/DAv10195/submit_server/db"
	"net/http"
)

const logHttpErrFormat = "error serving http request for %s"

func _stringForResp(e interface{}) string {
	errRespBytes, _ := json.Marshal(e)
	return string(errRespBytes)
}

type Response struct {
	Message	string	`json:"message"`
}

func (e *Response) String() string {
	return _stringForResp(e)
}

func writeResponse(w http.ResponseWriter, r *http.Request, httpStatus int, stringer fmt.Stringer) {
	w.WriteHeader(httpStatus)
	if _, err := w.Write([]byte(stringer.String())); err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
	}
}

func writeErrResp(w http.ResponseWriter, r *http.Request, httpStatus int, err error) {
	logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
	writeResponse(w, r, httpStatus, &Response{Message: err.Error()})
}

func writeStrErrResp(w http.ResponseWriter, r *http.Request, httpStatus int, str string) {
	err := fmt.Errorf(str)
	logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
	writeResponse(w, r, httpStatus, &Response{Message: err.Error()})
}

func writeElem(w http.ResponseWriter, r *http.Request, httpStatus int, e db.IBucketElement) {
	elemBytes, err := json.Marshal(e)
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(httpStatus)
	if _, err = w.Write(elemBytes); err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
	}
}

func writeElements(w http.ResponseWriter, r *http.Request, httpStatus int, elements []db.IBucketElement) {
	var elementsWrapper struct {
		Elements []db.IBucketElement `json:"elements"`
	}
	elementsWrapper.Elements = elements
	elementsBytes, err := json.Marshal(elementsWrapper)
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(httpStatus)
	if _, err = w.Write(elementsBytes); err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
	}
}
