package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/assignments"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/fs"
	"github.com/gorilla/mux"
	"io/ioutil"
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

func handleGetAssigmentDefsForCourse(forCourse string, w http.ResponseWriter, r *http.Request, params *submithttp.PagingParams) {
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.AssignmentDefinitions), func(_, elementBytes []byte) error {
		ass := &assignments.AssignmentDef{}
		if err := json.Unmarshal(elementBytes, ass); err != nil {
			return err
		}
		if ass.Course == forCourse {
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
	forCourse := r.Header.Get(submithttp.ForSubmitCourse)
	if forCourse != "" {
		handleGetAssigmentDefsForCourse(forCourse, w, r, params)
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
	if _, err := assignments.NewDef(ass.Course, ass.DueBy, ass.Name, r.Context().Value(authenticatedUser).(*users.User).UserName, true, fs.GetClient() != nil); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "assignment def created successfully"})
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
	updatedAss.Course = preUpdateAss.Course
	updatedAss.State = preUpdateAss.State
	updatedAss.Name = preUpdateAss.Name
	updatedAss.CreatedOn = preUpdateAss.CreatedOn
	updatedAss.CreatedBy = preUpdateAss.CreatedBy
	var elementsToUpdate []db.IBucketElement
	if updatedAss.DueBy != preUpdateAss.DueBy {
		if err := db.QueryBucket([]byte(db.AssignmentInstances), func(_ []byte, assInstBytes []byte) error {
			assInst := &assignments.AssignmentInstance{}
			if err := json.Unmarshal(assInstBytes, assInst); err != nil {
				return err
			}
			if assInst.DueBy == preUpdateAss.DueBy {
				elementsToUpdate = append(elementsToUpdate, assInst)
			}
			return nil
		}); err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	for _, assInstElem := range elementsToUpdate {
		assInstElem.(*assignments.AssignmentInstance).DueBy = updatedAss.DueBy
	}
	elementsToUpdate = append(elementsToUpdate, updatedAss)
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, elementsToUpdate...); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: fmt.Sprintf("assignment def '%s' updated successfully", updatedAss.Name)})
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
	if err := assignments.DeleteDef(ass, fs.GetClient() != nil); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusOK, &Response{Message: fmt.Sprintf("assignment def '%s' deleted successfully", ass.Name)})
}

func handlePublishAssignmentDef(w http.ResponseWriter, r *http.Request) {
	assKey, err := getAssDefKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	assDef, err := assignments.GetDef(assKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	if assDef.State == assignments.Published {
		writeStrErrResp(w, r, http.StatusBadRequest, fmt.Sprintf("assignment def '%s' already published", assDef.Name))
		return
	}
	var courseUserNames []string
	if err := db.QueryBucket([]byte(db.Users), func(_ []byte, userBytes []byte) error {
		user := &users.User{}
		if err := json.Unmarshal(userBytes, user); err != nil {
			return err
		}
		if user.CoursesAsStudent.Contains(assDef.Course) {
			courseUserNames = append(courseUserNames, user.UserName)
		}
		return nil
	}); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	var elements []db.IBucketElement
	asUser := r.Context().Value(authenticatedUser).(*users.User).UserName
	for _, userName := range courseUserNames {
		ass, err := assignments.NewInstance(assDef.Course, assDef.DueBy, assDef.Name, userName, asUser, false, fs.GetClient() != nil)
		if err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
		elements = append(elements, ass)
	}
	assDef.State = assignments.Published
	elements = append(elements, assDef)
	if err := db.Update(asUser, elements...); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusOK, &Response{Message: fmt.Sprintf("assignment def '%s' published successfully", assDef.Name)})
}

func initAssDefsRouter(r *mux.Router, manager *authManager) {
	basePath := fmt.Sprintf("/%s", db.AssignmentDefinitions)
	router := r.PathPrefix(basePath).Subrouter()
	router.HandleFunc("/", handleGetAssignmentDefs).Methods(http.MethodGet)
	router.HandleFunc("/", handleCreateAssignmentDef).Methods(http.MethodPost)
	manager.addPathToMap(fmt.Sprintf("%s/", basePath), func (user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		if request.Method == http.MethodGet {
			forCourse := request.Header.Get(submithttp.ForSubmitCourse)
			if forCourse != "" {
				return user.CoursesAsStaff.Contains(forCourse)
			}
		} else if request.Method == http.MethodPost {
			buf, err := ioutil.ReadAll(request.Body)
			if err != nil {
				return true // let the creation handler fail this with bad request...
			}
			bodyCopy1, bodyCopy2 := ioutil.NopCloser(bytes.NewBuffer(buf)), ioutil.NopCloser(bytes.NewBuffer(buf))
			request.Body = bodyCopy1
			ass := &assignments.AssignmentDef{}
			if err := json.NewDecoder(bodyCopy2).Decode(ass); err != nil {
				return true // let the creation handler fail this with bad request...
			}
			return user.CoursesAsStaff.Contains(ass.Course)
		}
		return false
	})
	specificPath := fmt.Sprintf("/{%s}/{%s}/{%s}", courseNumber, courseYear, assDefName)
	router.HandleFunc(specificPath, handleGetAssignmentDef).Methods(http.MethodGet)
	router.HandleFunc(specificPath, handleDeleteAssignmentDef).Methods(http.MethodDelete)
	router.HandleFunc(specificPath, handleUpdateAssignmentDef).Methods(http.MethodPut)
	router.HandleFunc(specificPath, handlePublishAssignmentDef).Methods(http.MethodPatch)
	manager.addRegex(regexp.MustCompile(fmt.Sprintf("^%s/.", basePath)), func (user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		cNumber, cYear, err := getCourseNumberAndYearFromRequest(request)
		if err != nil {
			return true // let handlers manage the error
		}
		courseKey := fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear)
		exists, err := db.KeyExistsInBucket([]byte(db.Courses), []byte(courseKey))
		if err != nil || !exists {
			return true // let the next handler send an appropriate error message
		}
		return user.CoursesAsStaff.Contains(courseKey)
	})
}
