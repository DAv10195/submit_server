package server

import (
	"encoding/json"
	"errors"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/courses"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/fs"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
	"strconv"
)

func getCourseNumberAndYearFromRequest(r *http.Request) (int, int, error) {
	number, err := strconv.ParseInt(mux.Vars(r)[courseNumber], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	year, err := strconv.ParseInt(mux.Vars(r)[courseYear], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return int(number), int(year), nil
}

func handleGetCourse(w http.ResponseWriter, r *http.Request) {
	number, year, err := getCourseNumberAndYearFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	course, err := courses.Get(fmt.Sprintf("%d%s%d", number, db.KeySeparator, year))
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	writeElem(w, r, http.StatusOK, course)
}

func handleGetCoursesForUser(forUser string, w http.ResponseWriter, r *http.Request, params *submithttp.PagingParams) {
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
	if err := db.QueryBucket([]byte(db.Courses), func(_, elementBytes []byte) error {
		course := &courses.Course{}
		if err := json.Unmarshal(elementBytes, course); err != nil {
			return err
		}
		if user.CoursesAsStudent.Contains(string(course.Key())) || user.CoursesAsStaff.Contains(string(course.Key())) {
			elementsIndex++
			if elementsIndex <= params.AfterId {
				return nil
			}
			elements = append(elements, course)
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

func handleGetCourses(w http.ResponseWriter, r *http.Request) {
	params, err := submithttp.PagingParamsFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, fmt.Errorf("error parsing query params: %v", err))
		return
	}
	forUser := r.Header.Get(submithttp.ForSubmitUser)
	if forUser != "" {
		handleGetCoursesForUser(forUser, w, r, params)
		return
	}
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.Courses), func(_, elementBytes []byte) error {
		elementsIndex++
		if elementsIndex <= params.AfterId {
			return nil
		}
		course := &courses.Course{}
		if err := json.Unmarshal(elementBytes, course); err != nil {
			return err
		}
		elements = append(elements, course)
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

func handleCreateCourse(w http.ResponseWriter, r *http.Request) {
	course := &courses.Course{}
	if err := json.NewDecoder(r.Body).Decode(course); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	if _, err := courses.NewCourse(course.Number, course.Name, r.Context().Value(authenticatedUser).(*users.User).UserName, true, fs.GetClient() != nil); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "course created successfully"})
}

func handleUpdateCourse(w http.ResponseWriter, r *http.Request) {
	number, year, err := getCourseNumberAndYearFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	courseKey := fmt.Sprintf("%d%s%d", number, db.KeySeparator, year)
	exists, err := db.KeyExistsInBucket([]byte(db.Courses), []byte(courseKey))
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	if !exists {
		writeStrErrResp(w, r, http.StatusNotFound, fmt.Sprintf("course '%s' doesn't exist", courseKey))
		return
	}
	updatedCourse := &courses.Course{}
	if err := json.NewDecoder(r.Body).Decode(updatedCourse); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	preUpdateCourse, err := courses.Get(courseKey)
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	updatedCourse.Number = preUpdateCourse.Number
	updatedCourse.Year = preUpdateCourse.Year
	updatedCourse.CreatedOn = preUpdateCourse.CreatedOn
	updatedCourse.CreatedBy = preUpdateCourse.CreatedBy
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, updatedCourse); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: fmt.Sprintf("course '%s' updated successfully", courseKey)})
}

func handleDeleteCourse(w http.ResponseWriter, r *http.Request) {
	number, year, err := getCourseNumberAndYearFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	course, err := courses.Get(fmt.Sprintf("%d%s%d", number, db.KeySeparator, year))
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	if err := courses.Delete(course, fs.GetClient() != nil); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusOK, &Response{Message: fmt.Sprintf("course '%d%s%d' deleted successfully", number, db.KeySeparator, year)})
}

// configure the courses router
func initCoursesRouter(r *mux.Router, manager *authManager) {
	coursesBasePath := fmt.Sprintf("/%s", db.Courses)
	coursesRouter := r.PathPrefix(coursesBasePath).Subrouter()
	coursesRouter.HandleFunc("/", handleGetCourses).Methods(http.MethodGet)
	coursesRouter.HandleFunc("/", handleCreateCourse).Methods(http.MethodPost)
	manager.addPathToMap(fmt.Sprintf("%s/", coursesBasePath), func (user *users.User, r *http.Request) bool {
		if user.Roles.Contains(users.Secretary) || user.Roles.Contains(users.Admin) {
			return true
		}
		return r.Method == http.MethodGet && user.Roles.Contains(users.StandardUser) && user.UserName == r.Header.Get(submithttp.ForSubmitUser)
	})
	specificCoursePath := fmt.Sprintf("/{%s}/{%s}", courseNumber, courseYear)
	coursesRouter.HandleFunc(specificCoursePath, handleGetCourse).Methods(http.MethodGet)
	coursesRouter.HandleFunc(specificCoursePath, handleDeleteCourse).Methods(http.MethodDelete)
	coursesRouter.HandleFunc(specificCoursePath, handleUpdateCourse).Methods(http.MethodPut)
	manager.addRegex(regexp.MustCompile(fmt.Sprintf("^%s/.", coursesBasePath)), func (user *users.User, r *http.Request) bool {
		if user.Roles.Contains(users.Admin) || user.Roles.Contains(users.Secretary) {
			return true
		}
		if user.Roles.Contains(users.StandardUser) {
			number, year, err := getCourseNumberAndYearFromRequest(r)
			if err != nil {
				return true // let the next handler send an appropriate error message
			}
			courseKey := fmt.Sprintf("%d%s%d", number, db.KeySeparator, year)
			exists, err := db.KeyExistsInBucket([]byte(db.Courses), []byte(courseKey))
			if err != nil || !exists {
				return true // let the next handler send an appropriate error message
			}
			if user.CoursesAsStaff.Contains(courseKey) && (r.Method == http.MethodGet || r.Method == http.MethodPut) {
				return true
			}
			if user.CoursesAsStudent.Contains(courseKey) && r.Method == http.MethodGet {
				return true
			}
		}
		return false
	})
}
