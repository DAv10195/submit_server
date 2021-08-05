package server

import (
	"encoding/json"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/appeals"
	"github.com/DAv10195/submit_server/elements/messages"
	"github.com/DAv10195/submit_server/elements/tests"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
)

func handleGetMessageBoxes(w http.ResponseWriter, r *http.Request) {
	params, err := submithttp.PagingParamsFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, fmt.Errorf("error parsing query params: %v", err))
		return
	}
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.MessageBoxes), func (_ []byte, msgBoxBytes []byte) error {
		elementsIndex++
		if elementsIndex <= params.AfterId {
			return nil
		}
		msgBox := &messages.MessageBox{}
		if err := json.Unmarshal(msgBoxBytes, msgBox); err != nil {
			return err
		}
		elements = append(elements, msgBox)
		elementsCount++
		if elementsCount == params.Limit {
			return &db.ErrStopQuery{}
		}
		return nil
	}); err != nil {
		if _, ok := err.(*db.ErrElementsLeftToProcess); ok {
			w.Header().Set(submithttp.ElementsLeftToProcess, trueStr)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	writeElements(w, r, http.StatusOK, elements)
}

func handleGetMessageBoxForUser(w http.ResponseWriter, r *http.Request) {
	params, err := submithttp.PagingParamsFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, fmt.Errorf("error parsing query params: %v", err))
		return
	}
	user, err := users.Get(mux.Vars(r)[userName])
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	msgBox, err := messages.Get(user.MessageBox)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.Messages), func (msgKey []byte, msgBytes []byte) error {
		if msgBox.Messages.Contains(string(msgKey)) {
			msg := &messages.Message{}
			if err := json.Unmarshal(msgBytes, msg); err != nil {
				return err
			}
			elementsIndex++
			if elementsIndex <= params.AfterId {
				return nil
			}
			elements = append(elements, msg)
			elementsCount++
			if elementsCount == params.Limit {
				return &db.ErrStopQuery{}
			}
		}
		return nil
	}); err != nil {
		if _, ok := err.(*db.ErrElementsLeftToProcess); ok {
			w.Header().Set(submithttp.ElementsLeftToProcess, trueStr)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	writeElements(w, r, http.StatusOK, elements)
}

func handlePostMessageToUser(w http.ResponseWriter, r *http.Request) {
	user, err := users.Get(mux.Vars(r)[userName])
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	msg := &messages.Message{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	if _, _, err := messages.NewMessage(r.Context().Value(authenticatedUser).(*users.User).UserName, msg.Text, user.MessageBox, true); err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "message created successfully"})
}

func handleGetMessageBoxForAppeal(w http.ResponseWriter, r *http.Request) {
	params, err := submithttp.PagingParamsFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, fmt.Errorf("error parsing query params: %v", err))
		return
	}
	assKey, err := getAssInstKey(r)
	if err != nil {
		writeStrErrResp(w, r, http.StatusBadRequest, "invalid course number and/or year integer path params")
		return
	}
	appeal, err := appeals.Get(assKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	msgBox, err := messages.Get(appeal.MessageBox)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.Messages), func (msgKey []byte, msgBytes []byte) error {
		if msgBox.Messages.Contains(string(msgKey)) {
			msg := &messages.Message{}
			if err := json.Unmarshal(msgBytes, msg); err != nil {
				return err
			}
			elementsIndex++
			if elementsIndex <= params.AfterId {
				return nil
			}
			elements = append(elements, msg)
			elementsCount++
			if elementsCount == params.Limit {
				return &db.ErrStopQuery{}
			}
		}
		return nil
	}); err != nil {
		if _, ok := err.(*db.ErrElementsLeftToProcess); ok {
			w.Header().Set(submithttp.ElementsLeftToProcess, trueStr)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	writeElements(w, r, http.StatusOK, elements)
}

func handlePostMessageToAppeal(w http.ResponseWriter, r *http.Request) {
	assKey, err := getAssInstKey(r)
	if err != nil {
		writeStrErrResp(w, r, http.StatusBadRequest, "invalid course number and/or year integer path params")
		return
	}
	appeal, err := appeals.Get(assKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	msg := &messages.Message{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	if _, _, err := messages.NewMessage(r.Context().Value(authenticatedUser).(*users.User).UserName, msg.Text, appeal.MessageBox, true); err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "message created successfully"})
}

func handleGetMessageBoxForTest(w http.ResponseWriter, r *http.Request) {
	params, err := submithttp.PagingParamsFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, fmt.Errorf("error parsing query params: %v", err))
		return
	}
	testKey, err := getTestKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	test, err := tests.Get(testKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	msgBox, err := messages.Get(test.MessageBox)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.Messages), func (msgKey []byte, msgBytes []byte) error {
		if msgBox.Messages.Contains(string(msgKey)) {
			msg := &messages.Message{}
			if err := json.Unmarshal(msgBytes, msg); err != nil {
				return err
			}
			elementsIndex++
			if elementsIndex <= params.AfterId {
				return nil
			}
			elements = append(elements, msg)
			elementsCount++
			if elementsCount == params.Limit {
				return &db.ErrStopQuery{}
			}
		}
		return nil
	}); err != nil {
		if _, ok := err.(*db.ErrElementsLeftToProcess); ok {
			w.Header().Set(submithttp.ElementsLeftToProcess, trueStr)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	writeElements(w, r, http.StatusOK, elements)
}

func handlePostMessageToTest(w http.ResponseWriter, r *http.Request) {
	testKey, err := getTestKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	test, err := tests.Get(testKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	msg := &messages.Message{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	if _, _, err := messages.NewMessage(r.Context().Value(authenticatedUser).(*users.User).UserName, msg.Text, test.MessageBox, true); err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "message created successfully"})
}

func initMessagesRouter(r *mux.Router, m *authManager) {
	basePath := fmt.Sprintf("/%s", db.Messages)
	router := r.PathPrefix(basePath).Subrouter()
	router.HandleFunc("/", handleGetMessageBoxes).Methods(http.MethodGet)
	m.addPathToMap(fmt.Sprintf("%s/", basePath), func(user *users.User, request *http.Request) bool {
		return user.Roles.Contains(users.Admin)
	})
	specificUserPath := fmt.Sprintf("/%s/{%s}", db.Users, userName)
	router.HandleFunc(specificUserPath, handleGetMessageBoxForUser).Methods(http.MethodGet)
	router.HandleFunc(specificUserPath, handlePostMessageToUser).Methods(http.MethodPost)
	m.addRegex(regexp.MustCompile(fmt.Sprintf("^%s/%s/.", basePath, db.Users)), func (user *users.User, r *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		if r.Method == http.MethodGet {
			return mux.Vars(r)[userName] == user.UserName
		}
		return r.Method == http.MethodPost
	})
	specificAppealPath := fmt.Sprintf("/%s/{%s}/{%s}/{%s}/{%s}", db.Appeals, courseNumber, courseYear, assDefName, userName)
	router.HandleFunc(specificAppealPath, handleGetMessageBoxForAppeal).Methods(http.MethodGet)
	router.HandleFunc(specificAppealPath, handlePostMessageToAppeal).Methods(http.MethodPost)
	m.addRegex(regexp.MustCompile(fmt.Sprintf("^%s/%s/.", basePath, db.Appeals)), func (user *users.User, r *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		if mux.Vars(r)[userName] == user.UserName {
			return true
		}
		cNumber, cYear, err := getCourseNumberAndYearFromRequest(r)
		if err != nil {
			return true // let the next handler send an appropriate error message
		}
		return user.CoursesAsStaff.Contains(fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear))
	})
	specificTestPath := fmt.Sprintf(fmt.Sprintf("/%s/{%s}/{%s}/{%s}/{%s}", db.Tests, courseNumber, courseYear, assDefName, testName))
	router.HandleFunc(specificTestPath, handleGetMessageBoxForTest).Methods(http.MethodGet)
	router.HandleFunc(specificTestPath, handlePostMessageToTest).Methods(http.MethodPost)
	m.addRegex(regexp.MustCompile(fmt.Sprintf("^%s/%s/.", basePath, db.Tests)), func (user *users.User, r *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		cNumber, cYear, err := getCourseNumberAndYearFromRequest(r)
		if err != nil {
			return true // let the next handler send an appropriate error message
		}
		if user.CoursesAsStaff.Contains(fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear)) {
			return true
		}
		testKey, err := getTestKey(r)
		if err != nil {
			return true // let the next handler send an appropriate error message
		}
		test, err := tests.Get(testKey)
		if err != nil {
			return true // let the next handler send an appropriate error message
		}
		return test.CreatedBy == user.UserName
	})
}
