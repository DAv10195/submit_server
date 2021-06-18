package server

import (
	"encoding/json"
	"fmt"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/appeals"
	"github.com/DAv10195/submit_server/elements/assignments"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
	"strings"
)

func handleGetAppealsForCourse(forCourse string, w http.ResponseWriter, r *http.Request, params *submithttp.PagingParams) {
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.Appeals), func (appealKey []byte, appealBytes []byte) error {
		elementsIndex++
		if elementsIndex <= params.AfterId {
			return nil
		}
		appeal := &appeals.Appeal{}
		if err := json.Unmarshal(appealBytes, appeal); err != nil {
			return err
		}
		if strings.HasPrefix(string(appealKey), forCourse) {
			elements = append(elements, appeal)
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

func handleGetAppealsForAss(forAss string, w http.ResponseWriter, r *http.Request, params *submithttp.PagingParams) {
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.Appeals), func (appealKey []byte, appealBytes []byte) error {
		elementsIndex++
		if elementsIndex <= params.AfterId {
			return nil
		}
		appeal := &appeals.Appeal{}
		if err := json.Unmarshal(appealBytes, appeal); err != nil {
			return err
		}
		if strings.HasPrefix(string(appealKey), forAss) {
			elements = append(elements, appeal)
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

func handleGetAppeals(w http.ResponseWriter, r *http.Request) {
	params, err := submithttp.PagingParamsFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, fmt.Errorf("error parsing query params: %v", err))
		return
	}
	forCourse := r.Header.Get(submithttp.ForSubmitCourse)
	forAss := r.Header.Get(submithttp.ForSubmitAss)
	if forCourse != "" && forAss != "" {
		writeStrErrResp(w, r, http.StatusBadRequest, "can't get appeals for both course and assignment instance")
		return
	}
	if forCourse != "" {
		handleGetAppealsForCourse(forCourse, w, r, params)
		return
	}
	if forAss != "" {
		handleGetAppealsForAss(forAss, w, r, params)
		return
	}
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.Appeals), func (_ []byte, appealBytes []byte) error {
		elementsIndex++
		if elementsIndex <= params.AfterId {
			return nil
		}
		appeal := &appeals.Appeal{}
		if err := json.Unmarshal(appealBytes, appeal); err != nil {
			return err
		}
		elements = append(elements, appeal)
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

func handleGetAppeal(w http.ResponseWriter, r *http.Request) {
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
	writeElem(w, r, http.StatusOK, appeal)
}

func handleCreateAppeal(w http.ResponseWriter, r *http.Request) {
	forAss := r.Header.Get(submithttp.ForSubmitAss)
	if forAss == "" {
		writeStrErrResp(w, r, http.StatusBadRequest, fmt.Sprintf("no assignment instance given via '%s' header", submithttp.ForSubmitAss))
		return
	}
	_, err := appeals.New(forAss, r.Context().Value(authenticatedUser).(*users.User).UserName, true)
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{"appeal created successfully"})
}

func handleUpdateAppealState(w http.ResponseWriter, r *http.Request) {
	stateStr := strings.ToLower(r.Header.Get(submithttp.SubmitState))
	var state int
	switch stateStr {
		case submithttp.AppealStateOpen:
			state = appeals.Open
		case submithttp.AppealStateClosed:
			state = appeals.Closed
		default:
			writeStrErrResp(w, r, http.StatusBadRequest, fmt.Sprintf("missing, empty or invalid state '%s' header", submithttp.SubmitState))
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
	if appeal.State == state {
		writeStrErrResp(w, r, http.StatusBadRequest, fmt.Sprintf("appeal is already in the given state ('%s')", stateStr))
		return
	}
	appeal.State = state
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, appeal); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusOK, &Response{Message: "appeal state updated successfully"})
}

func initAppealsRouter(r *mux.Router, m *authManager) {
	basePath := fmt.Sprintf("/%s", db.Appeals)
	router := r.PathPrefix(basePath).Subrouter()
	router.HandleFunc("/", handleGetAppeals).Methods(http.MethodGet)
	router.HandleFunc("/", handleCreateAppeal).Methods(http.MethodPost)
	m.addPathToMap(fmt.Sprintf("%s/", basePath), func(user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		forCourse := request.Header.Get(submithttp.ForSubmitCourse)
		forAss := request.Header.Get(submithttp.ForSubmitAss)
		if request.Method == http.MethodGet {
			if forCourse != "" && forAss != "" {
				return true // let the list handler fail this with bad request...
			}
			if forCourse != "" {
				return user.CoursesAsStaff.Contains(forCourse)
			}
			if forAss != ""  {
				ass, err := assignments.GetDef(forAss)
				if err != nil {
					return true // let the list handler fail this with bad request...
				}
				return user.CoursesAsStaff.Contains(ass.Course)
			}
		} else if request.Method == http.MethodPost {
			ass, err := assignments.GetInstance(forAss)
			if err != nil {
				return true // next handler will handle
			}
			return ass.UserName == user.UserName
		}
		return false
	})
	specificPath := fmt.Sprintf("/{%s}/{%s}/{%s}/{%s}", courseNumber, courseYear, assDefName, userName)
	router.HandleFunc(specificPath, handleGetAppeal).Methods(http.MethodGet)
	router.HandleFunc(specificPath, handleUpdateAppealState).Methods(http.MethodPatch)
	m.addRegex(regexp.MustCompile(fmt.Sprintf("%s/.", basePath)), func(user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		cNumber, cYear, err := getCourseNumberAndYearFromRequest(request)
		if err != nil {
			return true // let the next handler send an appropriate error message
		}
		courseKey := fmt.Sprintf("%d%s%d", cNumber, db.KeySeparator, cYear)
		exists, err := db.KeyExistsInBucket([]byte(db.Courses), []byte(courseKey))
		if err != nil || !exists {
			return true // let the next handler send an appropriate error message
		}
		if user.CoursesAsStaff.Contains(courseKey) {
			return true
		}
		assKey, err := getAssInstKey(request)
		ass, err := assignments.GetInstance(assKey)
		if err != nil {
			return true // let the next handler send an appropriate error message
		}
		return user.UserName == ass.UserName
	})
}
