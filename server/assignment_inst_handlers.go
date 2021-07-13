package server

import (
	"encoding/json"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/assignments"
	"github.com/DAv10195/submit_server/elements/tests"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
	"time"
)

func getAssInstKey(r *http.Request) (string, error) {
	assDefKey, err := getAssDefKey(r)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s%s", assDefKey, db.KeySeparator, mux.Vars(r)[userName]), nil
}

func handleGetAssignmentInstsForUser(forUser string, w http.ResponseWriter, r *http.Request, params *submithttp.PagingParams) {
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.AssignmentInstances), func(_ []byte, assInstBytes []byte) error {
		ass := &assignments.AssignmentInstance{}
		if err := json.Unmarshal(assInstBytes, ass); err != nil {
			return err
		}
		if ass.UserName == forUser {
			elementsIndex++
			if elementsIndex <= params.AfterId {
				return nil
			}
			elements = append(elements, ass)
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

func handleGetAssignmentInstsForAss(forAss string, w http.ResponseWriter, r *http.Request, params *submithttp.PagingParams) {
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.AssignmentInstances), func(_ []byte, assInstBytes []byte) error {
		ass := &assignments.AssignmentInstance{}
		if err := json.Unmarshal(assInstBytes, ass); err != nil {
			return err
		}
		if ass.AssignmentDef == forAss {
			elementsIndex++
			if elementsIndex <= params.AfterId {
				return nil
			}
			elements = append(elements, ass)
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

func handleGetAssignmentInsts(w http.ResponseWriter, r *http.Request) {
	params, err := submithttp.PagingParamsFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, fmt.Errorf("error parsing query params: %v", err))
		return
	}
	forUser := r.Header.Get(submithttp.ForSubmitUser)
	forAss := r.Header.Get(submithttp.ForSubmitAss)
	if forUser != "" && forAss != "" {
		writeStrErrResp(w, r, http.StatusBadRequest, "can't get assignment instances for both user and assignment def")
		return
	}
	if forUser != "" {
		handleGetAssignmentInstsForUser(forUser, w, r, params)
		return
	}
	if forAss != "" {
		handleGetAssignmentInstsForAss(forAss, w, r, params)
		return
	}
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.AssignmentInstances), func (_ []byte, assInstBytes []byte) error {
		elementsIndex++
		if elementsIndex <= params.AfterId {
			return nil
		}
		ass := &assignments.AssignmentInstance{}
		if err := json.Unmarshal(assInstBytes, ass); err != nil {
			return err
		}
		elements = append(elements, ass)
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

func handleGetAssignmentInst(w http.ResponseWriter, r *http.Request) {
	assKey, err := getAssInstKey(r)
	if err != nil {
		writeStrErrResp(w, r, http.StatusBadRequest, "invalid course number and/or year integer path params")
		return
	}
	ass, err := assignments.GetInstance(assKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	writeElem(w, r, http.StatusOK, ass)
}

func handleUpdateAssignmentInst(w http.ResponseWriter, r *http.Request) {
	assKey, err := getAssInstKey(r)
	if err != nil {
		writeStrErrResp(w, r, http.StatusBadRequest, "invalid course number and/or year integer path params")
		return
	}
	preUpdateAss, err := assignments.GetInstance(assKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	updatedAss := &assignments.AssignmentInstance{}
	if err := json.NewDecoder(r.Body).Decode(updatedAss); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	updatedAss.UserName = preUpdateAss.UserName
	updatedAss.AssignmentDef = preUpdateAss.AssignmentDef
	updatedAss.State = preUpdateAss.State
	updatedAss.MarkedAsCopy = preUpdateAss.MarkedAsCopy
	updatedAss.CreatedOn = preUpdateAss.CreatedOn
	updatedAss.CreatedBy = preUpdateAss.CreatedBy
	cNumber, cYear, err := getCourseNumberAndYearFromRequest(r)
	if err != nil {
		writeStrErrResp(w, r, http.StatusBadRequest, "invalid course number and/or year integer path params")
		return
	}
	courseKey := fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear)
	requestUser := r.Context().Value(authenticatedUser).(*users.User)
	if !requestUser.CoursesAsStaff.Contains(courseKey) && !requestUser.Roles.Contains(users.Admin) {
		if updatedAss.MarkedAsCopy != preUpdateAss.MarkedAsCopy {
			writeStrErrResp(w, r, http.StatusBadRequest, "updating assignment instance copy flag is forbidden")
			return
		}
		if updatedAss.Grade != preUpdateAss.Grade {
			writeStrErrResp(w, r, http.StatusBadRequest, "updating assignment instance grade is forbidden")
			return
		}
		if updatedAss.DueBy != preUpdateAss.DueBy {
			writeStrErrResp(w, r, http.StatusBadRequest, "updating assignment instance due date is forbidden")
			return
		}
	}
	if err := db.Update(requestUser.UserName, updatedAss); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: fmt.Sprintf("assignment instance '%s' updated successfully", string(updatedAss.Key()))})
}

func handleSubmitAssignmentInst(w http.ResponseWriter, r *http.Request) {
	assKey, err := getAssInstKey(r)
	if err != nil {
		writeStrErrResp(w, r, http.StatusBadRequest, "invalid course number and/or year integer path params")
		return
	}
	assInst, err := assignments.GetInstance(assKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	if assInst.State == assignments.Submitted {
		writeStrErrResp(w, r, http.StatusBadRequest, fmt.Sprintf("assignment instance '%s' already submitted", string(assInst.Key())))
		return
	}
	if assInst.State == assignments.Graded {
		writeStrErrResp(w, r, http.StatusBadRequest, fmt.Sprintf("assignment instance '%s' already graded", string(assInst.Key())))
		return
	}
	if time.Now().UTC().After(assInst.DueBy) {
		writeStrErrResp(w, r, http.StatusBadRequest, fmt.Sprintf("assignment instance '%s' can't be submitted anymore", string(assInst.Key())))
		return
	}
	assInst.State = assignments.Submitted
	requestUserName := r.Context().Value(authenticatedUser).(*users.User).UserName
	if err := db.Update(requestUserName, assInst); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	var testsToRun []string
	if err := db.QueryBucket([]byte(db.Tests), func (_, testBytes []byte) error {
		test := &tests.Test{}
		if err := json.Unmarshal(testBytes, test); err != nil {
			return err
		}
		if test.RunsOn == tests.OnSubmit {
			testsToRun = append(testsToRun, string(test.Key()))
		}
		return nil
	}); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	for _, testToRun := range testsToRun {
		tr, err := NewTestRequest(testToRun, assKey, true)
		if err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
		task, err := tr.ToTask(requestUserName, false)
		if err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
		task.Labels[onSubmitExec] = true
		if err := db.Update(requestUserName, task); err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	writeResponse(w, r, http.StatusOK, &Response{Message: fmt.Sprintf("assignment instance '%s' submitted successfully", string(assInst.Key()))})
}

func initAssInstsRouter(r *mux.Router, manager *authManager) {
	basePath := fmt.Sprintf("/%s", db.AssignmentInstances)
	router := r.PathPrefix(basePath).Subrouter()
	router.HandleFunc("/", handleGetAssignmentInsts).Methods(http.MethodGet)
	manager.addPathToMap(fmt.Sprintf("%s/", basePath), func (user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		forUser := request.Header.Get(submithttp.ForSubmitUser)
		forAss := request.Header.Get(submithttp.ForSubmitAss)
		if forUser != "" && forAss != "" {
			return true // let the list handler fail this with bad request...
		}
		if forUser != "" {
			return forUser == user.UserName
		}
		if forAss != "" {
			ass, err := assignments.GetDef(forAss)
			if err != nil {
				return true // let the list handler fail this with bad request...
			}
			return user.CoursesAsStaff.Contains(ass.Course)
		}
		return false
	})
	specificPath := fmt.Sprintf("/{%s}/{%s}/{%s}/{%s}", courseNumber, courseYear, assDefName, userName)
	router.HandleFunc(specificPath, handleGetAssignmentInst).Methods(http.MethodGet)
	router.HandleFunc(specificPath, handleUpdateAssignmentInst).Methods(http.MethodPut)
	router.HandleFunc(specificPath, handleSubmitAssignmentInst).Methods(http.MethodPatch)
	manager.addRegex(regexp.MustCompile(fmt.Sprintf("%s/.", basePath)), func (user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		cNumber, cYear, err := getCourseNumberAndYearFromRequest(request)
		if err != nil {
			return true // let handlers manage the error
		}
		courseKey := fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear)
		if user.CoursesAsStaff.Contains(courseKey) {
			return request.Method == http.MethodGet || request.Method == http.MethodPut
		}
		assKey, err := getAssInstKey(request)
		if err != nil {
			return true // next handler will handle
		}
		ass, err := assignments.GetInstance(assKey)
		if err != nil {
			return true // next handler will handle
		}
		return ass.UserName == user.UserName
	})
}
