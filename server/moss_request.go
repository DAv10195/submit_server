package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	commons "github.com/DAv10195/submit_commons"
	"github.com/DAv10195/submit_commons/containers"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/agents"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type MossRequest struct {
	AssignmentDef		string					`json:"assignment_def"`
	Users				*containers.StringSet	`json:"users"`
	Sensitivity			int						`json:"sensitivity"`
	Threshold			int						`json:"percentage"`
	Language			string					`json:"language"`
	ExecTimeout			int						`json:"timeout"`
}

// convert the copy detection request to a task to be executed by some agent
func (mr *MossRequest) ToTask(asUser string, withDbUpdate bool) (*agents.Task, error) {
	exists, err := db.KeyExistsInBucket([]byte(db.AssignmentDefinitions), []byte(mr.AssignmentDef))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &db.ErrKeyNotFoundInBucket{Bucket: db.AssignmentDefinitions, Key: mr.AssignmentDef}
	}
	if mr.Users == nil || mr.Users.NumberOfElements() < 2 {
		return nil, errors.New("request must have at least 2 users")
	}
	tb := agents.NewTaskBuilder(asUser, withDbUpdate)
	for _, username := range mr.Users.Slice() {
		assInstKey := fmt.Sprintf("%s%s%s", mr.AssignmentDef, db.KeySeparator, username)
		exists, err := db.KeyExistsInBucket([]byte(db.AssignmentInstances), []byte(assInstKey))
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, &db.ErrKeyNotFoundInBucket{Bucket: db.AssignmentInstances, Key: assInstKey}
		}
		tb.WithDependencies(fmt.Sprintf("/%s/%s/%s", db.Courses, strings.ReplaceAll(mr.AssignmentDef, db.KeySeparator, "/"), username))
	}
	if mr.Sensitivity < 1 || mr.Sensitivity > 1000 {
		return nil, errors.New("request sensitivity must be >= 1 ^ <= 1000")
	}
	if mr.Threshold < 1 || mr.Threshold > 100 {
		return nil, errors.New("request sensitivity must be >= 1 ^ <= 100")
	}
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s -l %s -m %d -d", commons.MossPathPlaceHolder, mr.Language, mr.Sensitivity))
	for _, username := range mr.Users.Slice() {
		sb.WriteString(fmt.Sprintf(" %s/*", username))
	}
	t, err := tb.WithCommand(sb.String()).WithResponseHandler(commons.Moss).WithExecTimeout(mr.ExecTimeout).
		WithLabel(commons.MossLink, "").WithLabel(assDefName, mr.AssignmentDef).WithLabel(commons.ExtractPaths, true).
		WithLabel(mossCopyThreshold, mr.Threshold).Build()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func handlePostMossRequest(w http.ResponseWriter, r *http.Request) {
	mr := &MossRequest{}
	if err := json.NewDecoder(r.Body).Decode(mr); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	task, err := mr.ToTask(r.Context().Value(authenticatedUser).(*users.User).UserName, true)
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &ResponseWithTaskId{Message: "task created successfully", TaskId: task.ID})
}

func handleGetMossResponse(w http.ResponseWriter, r *http.Request) {
	handleGetTaskResponse(w, r)
}

func initMossRequestRouter(r *mux.Router, m *authManager) {
	basePath := "/moss_requests"
	router := r.PathPrefix(basePath).Subrouter()
	router.HandleFunc("/", handlePostMossRequest).Methods(http.MethodPost)
	m.addPathToMap(fmt.Sprintf("%s/", basePath), func (user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		buf, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return true // let the next handler fail this with bad request...
		}
		bodyCopy1, bodyCopy2 := ioutil.NopCloser(bytes.NewBuffer(buf)), ioutil.NopCloser(bytes.NewBuffer(buf))
		request.Body = bodyCopy1
		mr := &MossRequest{}
		if err := json.NewDecoder(bodyCopy2).Decode(mr); err != nil {
			return true // let the next handler fail this with bad request...
		}
		split := strings.Split(mr.AssignmentDef, db.KeySeparator)
		if len(split) != 3 {
			return true // let the next handler fail this with bad request...
		}
		return user.CoursesAsStaff.Contains(fmt.Sprintf("%s:%s", split[0], split[1]))
	})
	router.HandleFunc(fmt.Sprintf("/{%s}", taskId), handleGetMossResponse).Methods(http.MethodGet)
	m.addRegex(regexp.MustCompile(fmt.Sprintf("%s/.", basePath)), func (user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		task, err := agents.GetTask(mux.Vars(request)[taskId])
		if err != nil {
			return true // let the next handler fail this request...
		}
		adn, ok := task.Labels[assDefName]
		if !ok {
			return true // let the next handler fail this request...
		}
		assDef, ok := adn.(string)
		if !ok {
			return true // let the next handler fail this request...
		}
		split := strings.Split(assDef, db.KeySeparator)
		if len(split) != 3 {
			return true // let the next handler fail this with bad request...
		}
		return user.CoursesAsStaff.Contains(fmt.Sprintf("%s:%s", split[0], split[1]))
	})
}
