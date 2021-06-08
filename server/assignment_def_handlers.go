package server

import (
	"encoding/json"
	"errors"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/assignments"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func getAssDefKey(r *http.Request) (string, error) {
	number, year, err := getCourseNumberAndYearFromRequest(r)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{strconv.Itoa(number), strconv.Itoa(year), mux.Vars(r)[assDefName]}, db.KeySeparator), nil
}

func handleGetAssigmentDefsForUser(forUser string, w http.ResponseWriter, r *http.Request, params *submithttp.PagingParams) {
	var user *users.User
	requestUser := r.Context().Value(authenticatedUser).(*users.User)
	if requestUser.UserName == forUser {
		user = requestUser
	} else {
		var err error
		user, err = users.Get(forUser)
		if err != nil {
			if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
				writeErrResp(w, r, http.StatusNotFound, err)
			} else {
				writeErrResp(w, r, http.StatusInternalServerError, err)
			}
			return
		}
	}
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.AssignmentDefinitions), func(_, elementBytes []byte) error {
		elementsIndex++
		if elementsIndex <= params.AfterId {
			return nil
		}
		ass := &assignments.AssignmentDef{}
		if err := json.Unmarshal(elementBytes, ass); err != nil {
			return err
		}
		if user.CoursesAsStudent.Contains(ass.Course) || user.Roles.Contains(ass.Course) {
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

func handleGetAssignmentDefs(w http.ResponseWriter, r *http.Request) {
	params, err := submithttp.PagingParamsFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, fmt.Errorf("error parsing query params: %v", err))
		return
	}
	forUser := r.Header.Get(submithttp.ForSubmitUser)
	if forUser != "" {
		handleGetAssigmentDefsForUser(forUser, w, r, params)
		return
	}
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.AssignmentDefinitions), func(_, elementBytes []byte) error {
		elementsIndex++
		if elementsIndex <= params.AfterId {
			return nil
		}
		ass := &assignments.AssignmentDef{}
		if err := json.Unmarshal(elementBytes, ass); err != nil {
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

func handleGetAssignmentDef(w http.ResponseWriter, r *http.Request) {
	assKey, err := getAssDefKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	ass, err := assignments.GetDef(assKey)
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

func handleCreateAssignmentDef(w http.ResponseWriter, r *http.Request) {
	ass := &assignments.AssignmentDef{}
	if err := json.NewDecoder(r.Body).Decode(ass); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	if _, err := assignments.NewDef(ass.Course, ass.DueBy, ass.Name, r.Context().Value(authenticatedUser).(*users.User).UserName, true); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{"assignment def created successfully"})
}

func handleUpdateAssignmentDef(w http.ResponseWriter, r *http.Request) {
	assKey, err := getAssDefKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	preUpdateAss, err := assignments.GetDef(assKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	updatedAss := &assignments.AssignmentDef{}
	if err := json.NewDecoder(r.Body).Decode(updatedAss); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	if updatedAss.Course != preUpdateAss.Course {
		writeStrErrResp(w, r, http.StatusBadRequest, "updating assignment def course is forbidden")
		return
	}
	if updatedAss.State != preUpdateAss.State {
		writeStrErrResp(w, r, http.StatusBadRequest, "updating assignment def state is forbidden")
		return
	}
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, updatedAss); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{fmt.Sprintf("assignment def '%s' updated successfully", updatedAss.Name)})
}

func handleDeleteAssignmentDef(w http.ResponseWriter, r *http.Request) {
	assKey, err := getAssDefKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	ass, err := assignments.GetDef(assKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	if err := assignments.DeleteDef(ass); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusOK, &Response{fmt.Sprintf("assignment def '%s' deleted successfully", ass.Name)})
}

func initAssDefsRouter(r *mux.Router, manager *authManager) {
	basePath := fmt.Sprintf("/%s", db.AssignmentDefinitions)
	router := r.PathPrefix(basePath).Subrouter()
	router.HandleFunc("/", handleGetAssignmentDefs).Methods(http.MethodGet)
	router.HandleFunc("/", handleCreateAssignmentDef).Methods(http.MethodPost)
	assAuthFunc := func (user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		cNumber, cYear, err := getCourseNumberAndYearFromRequest(request)
		if err != nil {
			logger.WithError(err).Errorf("error serving request for %s", request.URL.Path)
			return false
		}
		courseKey := fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear)
		return  user.CoursesAsStaff.Contains(courseKey) || (user.CoursesAsStudent.Contains(courseKey) && request.Method == http.MethodGet)
	}
	manager.addPathToMap(fmt.Sprintf("/%s", basePath), assAuthFunc)
	specificPath := fmt.Sprintf("/{%s}/{%s}/{%s}", courseNumber, courseYear, assDefName)
	router.HandleFunc(specificPath, handleGetAssignmentDef).Methods(http.MethodGet)
	router.HandleFunc(specificPath, handleDeleteAssignmentDef).Methods(http.MethodDelete)
	router.HandleFunc(specificPath, handleUpdateAssignmentDef).Methods(http.MethodPut)
	manager.addRegex(regexp.MustCompile(fmt.Sprintf("%s/.", basePath)), assAuthFunc)
}
