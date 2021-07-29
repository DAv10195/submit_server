package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/assignments"
	"github.com/DAv10195/submit_server/elements/tests"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/fs"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func getTestKey(r *http.Request) (string, error) {
	number, year, err := getCourseNumberAndYearFromRequest(r)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{strconv.Itoa(number), strconv.Itoa(year), mux.Vars(r)[assDefName], mux.Vars(r)[testName]}, db.KeySeparator), nil
}

func getTestsForUserAssignment(forUser, forAss string, w http.ResponseWriter, r *http.Request, params *submithttp.PagingParams) {
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.Tests), func (_ []byte, testBytes []byte) error {
		test := &tests.Test{}
		if err := json.Unmarshal(testBytes, test); err != nil {
			return err
		}
		if test.AssignmentDef == forAss && (test.CreatedBy == forUser || test.State == tests.Published) {
			elementsIndex++
			if elementsIndex <= params.AfterId {
				return nil
			}
			elements = append(elements, test)
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

func getTestsForAssignmentDef(forAss string, w http.ResponseWriter, r *http.Request, params *submithttp.PagingParams) {
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.Tests), func (_ []byte, testBytes []byte) error {
		test := &tests.Test{}
		if err := json.Unmarshal(testBytes, test); err != nil {
			return err
		}
		if test.AssignmentDef == forAss {
			elementsIndex++
			if elementsIndex <= params.AfterId {
				return nil
			}
			elements = append(elements, test)
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

func handleGetTests(w http.ResponseWriter, r *http.Request) {
	params, err := submithttp.PagingParamsFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, fmt.Errorf("error parsing query params: %v", err))
		return
	}
	forAss, forUser := r.Header.Get(submithttp.ForSubmitAss), r.Header.Get(submithttp.ForSubmitUser)
	if forAss != "" {
		if forUser != "" {
			getTestsForUserAssignment(forUser, forAss, w, r, params)
			return
		}
		getTestsForAssignmentDef(forAss, w, r, params)
		return
	}
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.Tests), func (_ []byte, testBytes []byte) error {
		elementsIndex++
		if elementsIndex <= params.AfterId {
			return nil
		}
		test := &tests.Test{}
		if err := json.Unmarshal(testBytes, test); err != nil {
			return err
		}
		elements = append(elements, test)
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

func handleCreateTest(w http.ResponseWriter, r *http.Request) {
	test := &tests.Test{}
	if err := json.NewDecoder(r.Body).Decode(test); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	if _, err := tests.New(r.Context().Value(authenticatedUser).(*users.User).UserName, test.AssignmentDef, test.Name, test.Command, test.OsType, test.Architecture, test.ExecTimeout, test.RunsOn, true, fs.GetClient() != nil); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "test created successfully"})
}

func handleGetTest(w http.ResponseWriter, r *http.Request) {
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
	writeElem(w, r, http.StatusOK, test)
}

func handleUpdateTest(w http.ResponseWriter, r *http.Request) {
	testKey, err := getTestKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	preUpdateTest, err := tests.Get(testKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	updatedTest := &tests.Test{}
	if err := json.NewDecoder(r.Body).Decode(updatedTest); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	updatedTest.Name = preUpdateTest.Name
	updatedTest.State = preUpdateTest.State
	updatedTest.AssignmentDef = preUpdateTest.AssignmentDef
	updatedTest.MessageBox = preUpdateTest.MessageBox
	updatedTest.CreatedOn = preUpdateTest.CreatedOn
	updatedTest.CreatedBy = preUpdateTest.CreatedBy
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, updatedTest); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: fmt.Sprintf("test '%s' updated successfully", updatedTest.Name)})
}

func handleDeleteTest(w http.ResponseWriter, r *http.Request) {
	testKey, err := getTestKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
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
	if err := tests.Delete(test, fs.GetClient() != nil); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusOK, &Response{Message: fmt.Sprintf("test '%s' deleted successfully", test.Name)})
}

func handleUpdateTestState(w http.ResponseWriter, r *http.Request) {
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
	switch test.State {
		case tests.Draft:
			test.State = tests.InReview
		case tests.InReview:
			test.State = tests.Published
		case tests.Published:
			writeStrErrResp(w, r, http.StatusBadRequest, "test is already published")
			return
		default:
			writeStrErrResp(w, r, http.StatusInternalServerError, "test state has invalid value")
			return
	}
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, test); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusOK, &Response{Message: fmt.Sprintf("test '%s' state updated successfully", string(test.Key()))})
}

func initTestsRouter(r *mux.Router, m *authManager) {
	basePath := fmt.Sprintf("/%s", db.Tests)
	router := r.PathPrefix(basePath).Subrouter()
	router.HandleFunc("/", handleGetTests).Methods(http.MethodGet)
	router.HandleFunc("/", handleCreateTest).Methods(http.MethodPost)
	m.addPathToMap(fmt.Sprintf("%s/", basePath), func(user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		if request.Method == http.MethodGet {
			forAss := request.Header.Get(submithttp.ForSubmitAss)
			forUser := request.Header.Get(submithttp.ForSubmitUser)
			if forAss != "" {
				ass, err := assignments.GetDef(forAss)
				if err != nil {
					return true // let actual handler send an appropriate error message
				}
				if user.CoursesAsStaff.Contains(ass.Course) {
					return true
				}
				return forUser != "" && user.UserName == forUser
			}
		} else if request.Method == http.MethodPost {
			buf, err := ioutil.ReadAll(request.Body)
			if err != nil {
				return true // let the creation handler fail this with bad request...
			}
			bodyCopy1, bodyCopy2 := ioutil.NopCloser(bytes.NewBuffer(buf)), ioutil.NopCloser(bytes.NewBuffer(buf))
			request.Body = bodyCopy1
			test := &tests.Test{}
			if err := json.NewDecoder(bodyCopy2).Decode(test); err != nil {
				return true // let the creation handler fail this with bad request...
			}
			ass, err := assignments.GetDef(test.AssignmentDef)
			if err != nil {
				return true // let the creation handler fail this with bad request...
			}
			return user.CoursesAsStaff.Contains(ass.Course) || (user.CoursesAsStudent.Contains(ass.Course) && test.RunsOn != tests.OnSubmit)
		}
		return false
	})
	specificPath := fmt.Sprintf("/{%s}/{%s}/{%s}/{%s}", courseNumber, courseYear, assDefName, testName)
	router.HandleFunc(specificPath, handleGetTest).Methods(http.MethodGet)
	router.HandleFunc(specificPath, handleDeleteTest).Methods(http.MethodDelete)
	router.HandleFunc(specificPath, handleUpdateTest).Methods(http.MethodPut)
	router.HandleFunc(specificPath, handleUpdateTestState).Methods(http.MethodPatch)
	m.addRegex(regexp.MustCompile(fmt.Sprintf("^%s/.", basePath)), func (user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		testKey, err := getTestKey(request)
		if err != nil {
			return true // let the next handler send an appropriate error message
		}
		test, err := tests.Get(testKey)
		if err != nil {
			return true // let the next handler send an appropriate error message
		}
		cNumber, cYear, err := getCourseNumberAndYearFromRequest(request)
		if err != nil {
			return true // let the next handler send an appropriate error message
		}
		courseKey := fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear)
		if user.CoursesAsStaff.Contains(courseKey) {
			return true
		} else if user.CoursesAsStudent.Contains(courseKey) {
			if request.Method == http.MethodPatch && test.State == tests.InReview {
				return false
			}
			return test.CreatedBy == user.UserName || (test.State == tests.Published && request.Method == http.MethodGet)
		}
		return false
	})
}
