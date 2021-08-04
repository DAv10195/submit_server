package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/DAv10195/submit_commons/containers"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/agents"
	"github.com/DAv10195/submit_server/elements/assignments"
	"github.com/DAv10195/submit_server/elements/tests"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// test a single assignment instance or a definition if only a test is given
type TestRequest struct {
	Test					string		`json:"test"`
	AssignmentInstance		string		`json:"assignment_instance"`
	OnDemand				bool		`json:"on_demand"`
}

// test multiple assignment instances
type MultiTestRequest struct {
	Test					string					`json:"test"`
	AssignmentInstances		*containers.StringSet	`json:"assignment_instances"`
}

func NewTestRequest(test, assInst string, onDemand bool) (*TestRequest, error) {
	testObj, err := tests.Get(test)
	if err != nil {
		return nil, err
	}
	if assInst != "" {
		assInstObj, err := assignments.GetInstance(assInst)
		if err != nil {
			return nil, err
		}
		if assInstObj.AssignmentDef != testObj.AssignmentDef {
			return nil, fmt.Errorf("selected assignment instance ('%s') is unrelated to the selected test assignment def ('%s')", assInstObj.AssignmentDef, testObj.AssignmentDef )
		}
	}
	return &TestRequest{test, assInst, onDemand}, nil
}

// convert the test request to a task to be executed by an agent
func (tr *TestRequest) ToTask(asUser string, withDbUpdate bool) (*agents.Task, error) {
	testObj, err := tests.Get(tr.Test)
	if err != nil {
		return nil, err
	}
	tb := agents.NewTaskBuilder(asUser, withDbUpdate)
	tb.WithArchitecture(testObj.Architecture).WithOsType(testObj.OsType).WithCommand(testObj.Command).
		WithExecTimeout(testObj.ExecTimeout).WithResponseHandler(testTask).WithLabel(assDefName, testObj.AssignmentDef).
		WithLabel(testName, testObj.Name).WithLabel(userName, asUser)
	for _, testFile := range testObj.Files.Slice() {
		tb.WithDependencies(fmt.Sprintf("/%s/%s/tests/%s/%s", db.Courses, strings.ReplaceAll(testObj.AssignmentDef, db.KeySeparator, "/"), testObj.Name, testFile))
	}
	if tr.AssignmentInstance != "" {
		assInst, err := assignments.GetInstance(tr.AssignmentInstance)
		if err != nil {
			return nil, err
		}
		for _, assInstFile := range assInst.Files.Slice() {
			tb.WithDependencies(fmt.Sprintf("/%s/%s/%s/%s", db.Courses, strings.ReplaceAll(testObj.AssignmentDef, db.KeySeparator, "/"), assInst.UserName, assInstFile))
		}
		tb.WithLabel(assInstUsrName, assInst.UserName)
		tb.WithLabel(onDemandTask, tr.OnDemand)
	} else {
		assDef, err := assignments.GetDef(testObj.AssignmentDef)
		if err != nil {
			return nil, err
		}
		for _, assDefFile := range assDef.Files.Slice() {
			tb.WithDependencies(fmt.Sprintf("/%s/%s/%s", db.Courses, strings.ReplaceAll(testObj.AssignmentDef, db.KeySeparator, "/"), assDefFile))
		}
		tb.WithLabel(onDemandTask, true)
	}
	t, err := tb.Build()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func handlePostTestRequest(w http.ResponseWriter, r *http.Request) {
	tr := &TestRequest{}
	if err := json.NewDecoder(r.Body).Decode(tr); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	tr, err := NewTestRequest(tr.Test, tr.AssignmentInstance, tr.OnDemand)
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	task, err := tr.ToTask(r.Context().Value(authenticatedUser).(*users.User).UserName, true)
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &ResponseWithTaskId{Message: "task created successfully", TaskId: task.ID})
}

func handleGetTestResponse(w http.ResponseWriter, r *http.Request) {
	handleGetTaskResponse(w, r)
}

func handlePostMultiTestRequest(w http.ResponseWriter, r *http.Request) {
	mtr := &MultiTestRequest{}
	if err := json.NewDecoder(r.Body).Decode(mtr); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	if mtr.AssignmentInstances == nil || mtr.AssignmentInstances.NumberOfElements() == 0 {
		if mtr.AssignmentInstances == nil {
			mtr.AssignmentInstances = containers.NewStringSet()
		}
		test, err := tests.Get(mtr.Test)
		if err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
		if err := db.QueryBucket([]byte(db.AssignmentInstances), func (_, elemBytes []byte) error {
			ass := &assignments.AssignmentInstance{}
			if err := json.Unmarshal(elemBytes, ass); err != nil {
				return err
			}
			if ass.AssignmentDef == test.AssignmentDef {
				mtr.AssignmentInstances.Add(string(ass.Key()))
			}
			return nil
		}); err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	var notSubmittedAssInsts []db.IBucketElement
	for _, assInstKey := range mtr.AssignmentInstances.Slice() {
		assInst, err := assignments.GetInstance(assInstKey)
		if err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
		if assInst.State == assignments.Assigned {
			assInst.Grade = 0
			assInst.State = assignments.Graded
			notSubmittedAssInsts = append(notSubmittedAssInsts, assInst)
			continue
		}
		tr, err := NewTestRequest(mtr.Test, assInstKey, false)
		if err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
		if _, err := tr.ToTask(r.Context().Value(authenticatedUser).(*users.User).UserName, true); err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	if len(notSubmittedAssInsts) > 0 {
		if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, notSubmittedAssInsts...); err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "assignment is being checked"})
}

func initTestRequestsRouter(r *mux.Router, m *authManager) {
	basePath := "/test_requests"
	router := r.PathPrefix(basePath).Subrouter()
	router.HandleFunc("/single", handlePostTestRequest).Methods(http.MethodPost)
	m.addPathToMap(fmt.Sprintf("%s/single", basePath), func(user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		buf, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return true // let the next handler fail this with bad request...
		}
		bodyCopy1, bodyCopy2 := ioutil.NopCloser(bytes.NewBuffer(buf)), ioutil.NopCloser(bytes.NewBuffer(buf))
		request.Body = bodyCopy1
		tr := &TestRequest{}
		if err := json.NewDecoder(bodyCopy2).Decode(tr); err != nil {
			return true // let the next handler fail this with bad request...
		}
		test, err := tests.Get(tr.Test)
		if err != nil {
			return true // let the next handler fail this request...
		}
		assDef, err := assignments.GetDef(test.AssignmentDef)
		if err != nil {
			return true // let the next handler fail this request...
		}
		if tr.AssignmentInstance != "" {
			assInst, err := assignments.GetInstance(tr.AssignmentInstance)
			if err != nil {
				return true // let the next handler fail this request...
			}
			if assInst.AssignmentDef != test.AssignmentDef {
				 return false
			}
			if tr.OnDemand && assInst.UserName == user.UserName {
				return true
			}
		}
		return user.CoursesAsStaff.Contains(assDef.Course)
	})
	router.HandleFunc("/multi", handlePostMultiTestRequest).Methods(http.MethodPost)
	m.addPathToMap(fmt.Sprintf("%s/multi", basePath), func(user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		buf, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return true // let the next handler fail this with bad request...
		}
		bodyCopy1, bodyCopy2 := ioutil.NopCloser(bytes.NewBuffer(buf)), ioutil.NopCloser(bytes.NewBuffer(buf))
		request.Body = bodyCopy1
		mtr := &MultiTestRequest{}
		if err := json.NewDecoder(bodyCopy2).Decode(mtr); err != nil {
			return true // let the next handler fail this with bad request...
		}
		test, err := tests.Get(mtr.Test)
		if err != nil {
			return true // let the next handler fail this request...
		}
		assDef, err := assignments.GetDef(test.AssignmentDef)
		if err != nil {
			return true // let the next handler fail this request...
		}
		return user.CoursesAsStaff.Contains(assDef.Course)
	})
	router.HandleFunc(fmt.Sprintf("/{%s}", taskId), handleGetTestResponse).Methods(http.MethodGet)
	m.addRegex(regexp.MustCompile(fmt.Sprintf("^%s/.", basePath)), func (user *users.User, request *http.Request) bool {
		if user.Roles.Contains(users.Admin) {
			return true
		}
		task, err := agents.GetTask(mux.Vars(request)[taskId])
		if err != nil {
			return true // let the next handler fail this request...
		}
		od, ok := task.Labels[onDemandTask]
		if !ok {
			return false
		}
		onDemand, ok := od.(bool)
		if !ok {
			return false
		}
		return onDemand && task.CreatedBy == user.UserName
	})
}
