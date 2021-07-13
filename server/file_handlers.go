package server

import (
	"bytes"
	"errors"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/assignments"
	"github.com/DAv10195/submit_server/elements/courses"
	"github.com/DAv10195/submit_server/elements/tests"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/fs"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

func getFileNamesInRequest(r *http.Request) ([]string, error) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	bodyCopy1, bodyCopy2 := ioutil.NopCloser(bytes.NewBuffer(buf)), ioutil.NopCloser(bytes.NewBuffer(buf))
	r.Body = bodyCopy1
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	var fileNames []string
	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			fileNames = append(fileNames, fileHeader.Filename)
		}
	}
	r.Body = bodyCopy2
	return fileNames, nil
}

func handlePostFileForCourse(w http.ResponseWriter, r *http.Request) {
	cNumber, cYear, err := getCourseNumberAndYearFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	course, err := courses.Get(fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear))
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	fileNames, err := getFileNamesInRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	if err := fs.GetClient().ForwardBody(fmt.Sprintf("/%s/%d/%d", db.Courses, course.Number, course.Year), r.Header.Get(ContentType), r.Body); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	course.Files.Add(fileNames...)
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, course); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "file uploaded successfully"})
}

func handleGetFileForCourse(w http.ResponseWriter, r *http.Request) {
	cNumber, cYear, err := getCourseNumberAndYearFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	course, err := courses.Get(fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear))
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	fileName := r.Header.Get(submithttp.SubmitFile)
	if fileName == "" {
		writeStrErrResp(w, r, http.StatusBadRequest, "no file name given")
		return
	}
	if !course.Files.Contains(fileName) {
		writeStrErrResp(w, r, http.StatusNotFound, "file not found")
		return
	}
	writer := &bytes.Buffer{}
	respHeaders, err := fs.GetClient().DownloadFile(fmt.Sprintf("/%s/%d/%d/%s", db.Courses, cNumber, cYear, fileName), writer)
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	for k, v := range respHeaders {
		w.Header().Del(k)
		for _, hv := range v {
			w.Header().Add(k, hv)
		}
	}
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, writer); err != nil {
		logger.WithError(err).Error("error copying data from file server to client")
		return
	}
}

func handleDeleteFileForCourse(w http.ResponseWriter, r *http.Request) {
	cNumber, cYear, err := getCourseNumberAndYearFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	course, err := courses.Get(fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear))
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	fileName := r.Header.Get(submithttp.SubmitFile)
	if fileName == "" {
		writeStrErrResp(w, r, http.StatusBadRequest, "no file name given")
		return
	}
	if !course.Files.Contains(fileName) {
		writeStrErrResp(w, r, http.StatusNotFound, "file not found")
		return
	}
	if err := fs.GetClient().Delete(fmt.Sprintf("/%s/%d/%d/%s", db.Courses, cNumber, cYear, fileName)); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	course.Files.Remove(fileName)
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, course); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "file deleted successfully"})
}

func handlePostFileForAssignmentDef(w http.ResponseWriter, r *http.Request) {
	assDefKey, err := getAssDefKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	ass, err := assignments.GetDef(assDefKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	fileNames, err := getFileNamesInRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	vars := mux.Vars(r)
	if err := fs.GetClient().ForwardBody(fmt.Sprintf("/%s/%s/%s/%s", db.Courses, vars[courseNumber], vars[courseYear], ass.Name), r.Header.Get(ContentType), r.Body); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	ass.Files.Add(fileNames...)
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, ass); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "file uploaded successfully"})
}

func handleGetFileForAssignmentDef(w http.ResponseWriter, r *http.Request) {
	assDefKey, err := getAssDefKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	ass, err := assignments.GetDef(assDefKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	fileName := r.Header.Get(submithttp.SubmitFile)
	if fileName == "" {
		writeStrErrResp(w, r, http.StatusBadRequest, "no file name given")
		return
	}
	if !ass.Files.Contains(fileName) {
		writeStrErrResp(w, r, http.StatusNotFound, "file not found")
		return
	}
	vars := mux.Vars(r)
	writer := &bytes.Buffer{}
	respHeaders, err := fs.GetClient().DownloadFile(fmt.Sprintf("/%s/%s/%s/%s/%s", db.Courses, vars[courseNumber], vars[courseYear], ass.Name, fileName), writer)
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	for k, v := range respHeaders {
		w.Header().Del(k)
		for _, hv := range v {
			w.Header().Add(k, hv)
		}
	}
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, writer); err != nil {
		logger.WithError(err).Error("error copying data from file server to client")
		return
	}
}

func handleDeleteFileForAssignmentDef(w http.ResponseWriter, r *http.Request) {
	assDefKey, err := getAssDefKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	ass, err := assignments.GetDef(assDefKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	fileName := r.Header.Get(submithttp.SubmitFile)
	if fileName == "" {
		writeStrErrResp(w, r, http.StatusBadRequest, "no file name given")
		return
	}
	if !ass.Files.Contains(fileName) {
		writeStrErrResp(w, r, http.StatusNotFound, "file not found")
		return
	}
	vars := mux.Vars(r)
	if err := fs.GetClient().Delete(fmt.Sprintf("/%s/%s/%s/%s/%s", db.Courses, vars[courseNumber], vars[courseYear], ass.Name, fileName)); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	ass.Files.Remove(fileName)
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, ass); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "file deleted successfully"})
}

func handlePostFileForTest(w http.ResponseWriter, r *http.Request) {
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
	fileNames, err := getFileNamesInRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	vars := mux.Vars(r)
	if err := fs.GetClient().ForwardBody(fmt.Sprintf("/%s/%s/%s/%s/tests/%s", db.Courses, vars[courseNumber], vars[courseYear], vars[assDefName], test.Name), r.Header.Get(ContentType), r.Body); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	test.Files.Add(fileNames...)
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, test); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "file uploaded successfully"})
}

func handleGetFileForTest(w http.ResponseWriter, r *http.Request) {
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
	fileName := r.Header.Get(submithttp.SubmitFile)
	if fileName == "" {
		writeStrErrResp(w, r, http.StatusBadRequest, "no file name given")
		return
	}
	if !test.Files.Contains(fileName) {
		writeStrErrResp(w, r, http.StatusNotFound, "file not found")
		return
	}
	vars := mux.Vars(r)
	writer := &bytes.Buffer{}
	respHeaders, err := fs.GetClient().DownloadFile(fmt.Sprintf("/%s/%s/%s/%s/tests/%s/%s", db.Courses, vars[courseNumber], vars[courseYear], vars[assDefName], test.Name, fileName), writer)
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	for k, v := range respHeaders {
		w.Header().Del(k)
		for _, hv := range v {
			w.Header().Add(k, hv)
		}
	}
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, writer); err != nil {
		logger.WithError(err).Error("error copying data from file server to client")
		return
	}
}

func handleDeleteFileForTest(w http.ResponseWriter, r *http.Request) {
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
	fileName := r.Header.Get(submithttp.SubmitFile)
	if fileName == "" {
		writeStrErrResp(w, r, http.StatusBadRequest, "no file name given")
		return
	}
	if !test.Files.Contains(fileName) {
		writeStrErrResp(w, r, http.StatusNotFound, "file not found")
		return
	}
	vars := mux.Vars(r)
	if err := fs.GetClient().Delete(fmt.Sprintf("/%s/%s/%s/%s/tests/%s/%s", db.Courses, vars[courseNumber], vars[courseYear], vars[assDefName], test.Name, fileName)); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	test.Files.Remove(fileName)
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, test); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "file deleted successfully"})
}

func handlePostFileForAssignmentInst(w http.ResponseWriter, r *http.Request) {
	assInstKey, err := getAssInstKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	ass, err := assignments.GetInstance(assInstKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	if ass.State == assignments.Submitted {
		writeStrErrResp(w, r, http.StatusBadRequest, "can't upload file for a submitted assignment instance")
		return
	} else if ass.State == assignments.Graded {
		writeStrErrResp(w, r, http.StatusBadRequest, "can't upload file for a graded assignment instance")
		return
	}
	fileNames, err := getFileNamesInRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	vars := mux.Vars(r)
	if err := fs.GetClient().ForwardBody(fmt.Sprintf("/%s/%s/%s/%s/%s", db.Courses, vars[courseNumber], vars[courseYear], vars[assDefName], ass.UserName), r.Header.Get(ContentType), r.Body); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	ass.Files.Add(fileNames...)
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, ass); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "file uploaded successfully"})
}

func handleGetFileForAssignmentInst(w http.ResponseWriter, r *http.Request) {
	assInstKey, err := getAssInstKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	ass, err := assignments.GetInstance(assInstKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	fileName := r.Header.Get(submithttp.SubmitFile)
	if fileName == "" {
		writeStrErrResp(w, r, http.StatusBadRequest, "no file name given")
		return
	}
	if !ass.Files.Contains(fileName) {
		writeStrErrResp(w, r, http.StatusNotFound, "file not found")
		return
	}
	vars := mux.Vars(r)
	writer := &bytes.Buffer{}
	respHeaders, err := fs.GetClient().DownloadFile(fmt.Sprintf("/%s/%s/%s/%s/%s/%s", db.Courses, vars[courseNumber], vars[courseYear], vars[assDefName], ass.UserName, fileName), writer)
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	for k, v := range respHeaders {
		w.Header().Del(k)
		for _, hv := range v {
			w.Header().Add(k, hv)
		}
	}
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, writer); err != nil {
		logger.WithError(err).Error("error copying data from file server to client")
		return
	}
}

func handleDeleteFileForAssignmentInst(w http.ResponseWriter, r *http.Request) {
	assInstKey, err := getAssInstKey(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, errors.New("invalid course number and/or year integer path params"))
		return
	}
	ass, err := assignments.GetInstance(assInstKey)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	if ass.State == assignments.Submitted {
		writeStrErrResp(w, r, http.StatusBadRequest, "can't delete file for a submitted assignment instance")
		return
	} else if ass.State == assignments.Graded {
		writeStrErrResp(w, r, http.StatusBadRequest, "can't delete file for a graded assignment instance")
		return
	}
	fileName := r.Header.Get(submithttp.SubmitFile)
	if fileName == "" {
		writeStrErrResp(w, r, http.StatusBadRequest, "no file name given")
		return
	}
	if !ass.Files.Contains(fileName) {
		writeStrErrResp(w, r, http.StatusNotFound, "file not found")
		return
	}
	vars := mux.Vars(r)
	if err := fs.GetClient().Delete(fmt.Sprintf("/%s/%s/%s/%s/%s/%s", db.Courses, vars[courseNumber], vars[courseYear], vars[assDefName], ass.UserName, fileName)); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	ass.Files.Remove(fileName)
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, ass); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "file deleted successfully"})
}

func initFilesRouter(r *mux.Router, m *authManager) {
	router := r.PathPrefix("/files").Subrouter()
	specificCoursePath := fmt.Sprintf("/%s/{%s}/{%s}", db.Courses, courseNumber, courseYear)
	router.HandleFunc(specificCoursePath, handleGetFileForCourse).Methods(http.MethodGet)
	router.HandleFunc(specificCoursePath, handlePostFileForCourse).Methods(http.MethodPost)
	router.HandleFunc(specificCoursePath, handleDeleteFileForCourse).Methods(http.MethodDelete)
	m.addRegex(regexp.MustCompile(fmt.Sprintf("/files/%s/.", db.Courses)), func (user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) || user.Roles.Contains(users.Secretary) {
			return true
		}
		cNumber, cYear, err := getCourseNumberAndYearFromRequest(request)
		if err != nil {
			return true // let the next handler send an appropriate error message
		}
		courseKey := fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear)
		if request.Method == http.MethodGet {
			return user.CoursesAsStaff.Contains(courseKey) || user.CoursesAsStudent.Contains(courseKey)
		} else if request.Method == http.MethodPost || request.Method == http.MethodDelete {
			return user.CoursesAsStaff.Contains(courseKey)
		}
		return false
	})
	specificAssDefPath := fmt.Sprintf("/%s/{%s}/{%s}/{%s}", db.AssignmentDefinitions, courseNumber, courseYear, assDefName)
	router.HandleFunc(specificAssDefPath, handleGetFileForAssignmentDef).Methods(http.MethodGet)
	router.HandleFunc(specificAssDefPath, handlePostFileForAssignmentDef).Methods(http.MethodPost)
	router.HandleFunc(specificAssDefPath, handleDeleteFileForAssignmentDef).Methods(http.MethodDelete)
	m.addRegex(regexp.MustCompile(fmt.Sprintf("/files/%s/.", db.AssignmentDefinitions)), func (user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		cNumber, cYear, err := getCourseNumberAndYearFromRequest(request)
		if err != nil {
			return true // let the next handler send an appropriate error message
		}
		return user.CoursesAsStaff.Contains(fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear))
	})
	specificAssInstPath := fmt.Sprintf("/%s/{%s}/{%s}/{%s}/{%s}", db.AssignmentInstances, courseNumber, courseYear, assDefName, userName)
	router.HandleFunc(specificAssInstPath, handleGetFileForAssignmentInst).Methods(http.MethodGet)
	router.HandleFunc(specificAssInstPath, handlePostFileForAssignmentInst).Methods(http.MethodPost)
	router.HandleFunc(specificAssInstPath, handleDeleteFileForAssignmentInst).Methods(http.MethodDelete)
	m.addRegex(regexp.MustCompile(fmt.Sprintf("/files/%s/.", db.AssignmentInstances)), func (user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		if request.Method == http.MethodPost || request.Method == http.MethodDelete {
			assInstKey, err := getAssInstKey(request)
			if err != nil {
				return true // let the next handler send an appropriate error message
			}
			ass, err := assignments.GetInstance(assInstKey)
			if err != nil {
				return true // let the next handler send an appropriate error message
			}
			return mux.Vars(request)[userName] == user.UserName && !time.Now().UTC().After(ass.DueBy)
		}
		cNumber, cYear, err := getCourseNumberAndYearFromRequest(request)
		if err != nil {
			return true // let the next handler send an appropriate error message
		}
		return request.Method == http.MethodGet && user.CoursesAsStaff.Contains(fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear))
	})
	specificTestPath := fmt.Sprintf("/%s/{%s}/{%s}/{%s}/{%s}", db.Tests, courseNumber, courseYear, assDefName, testName)
	router.HandleFunc(specificTestPath, handleGetFileForTest).Methods(http.MethodGet)
	router.HandleFunc(specificTestPath, handlePostFileForTest).Methods(http.MethodPost)
	router.HandleFunc(specificTestPath, handleDeleteFileForTest).Methods(http.MethodDelete)
	m.addRegex(regexp.MustCompile(fmt.Sprintf("/files/%s/.", db.Tests)), func (user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		cNumber, cYear, err := getCourseNumberAndYearFromRequest(request)
		if err != nil {
			return true // let the next handler send an appropriate error message
		}
		courseKey := fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear)
		if user.CoursesAsStaff.Contains(courseKey) {
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
		return test.CreatedBy == user.UserName
	})
}
