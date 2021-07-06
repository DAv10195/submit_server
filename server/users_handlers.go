package server

import (
	"encoding/json"
	"fmt"
	"github.com/DAv10195/submit_commons/errors"
	submithttp "github.com/DAv10195/submit_commons/http"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/messages"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/fs"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
	"strings"
)

// return information about the requested user
func handleGetUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(authenticatedUser).(*users.User)
	requestedUserName := mux.Vars(r)[userName]
	var requestedUser *users.User
	if requestedUserName == user.UserName {
		requestedUser = user
	} else {
		var err error
		requestedUser, err = users.Get(requestedUserName)
		if err != nil {
			if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
				writeErrResp(w, r, http.StatusNotFound, err)
			} else {
				writeErrResp(w, r, http.StatusInternalServerError, err)
			}
			return
		}
	}
	writeElem(w, r, http.StatusOK, requestedUser)
}

// return information about all users
func handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	params, err := submithttp.PagingParamsFromRequest(r)
	if err != nil {
		writeErrResp(w, r, http.StatusBadRequest, fmt.Errorf("error parsing query params: %v", err))
		return
	}
	var elements []db.IBucketElement
	var elementsCount, elementsIndex int64
	if err := db.QueryBucket([]byte(db.Users), func(_, elementBytes []byte) error {
		elementsIndex++
		if elementsIndex <= params.AfterId {
			return nil
		}
		user := &users.User{}
		if err := json.Unmarshal(elementBytes, user); err != nil {
			return err
		}
		elements = append(elements, user)
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

// register the given users with their given information
func handleRegisterUsers(w http.ResponseWriter, r *http.Request) {
	requestUser := r.Context().Value(authenticatedUser).(*users.User)
	var body struct {
		Users	[]*users.User	`json:"users"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	var elementsToCreate []db.IBucketElement
	for _, u := range body.Users {
		builder := users.NewUserBuilder(requestUser.UserName, false)
		user, err := builder.WithUserName(u.UserName).WithFirstName(u.FirstName).WithLastName(u.LastName).
			WithPassword(u.Password).WithEmail(u.Email).WithRoles(u.Roles.Slice()...).
			WithCoursesAsStaff(u.CoursesAsStaff.Slice()...).WithCoursesAsStudent(u.CoursesAsStudent.Slice()...).Build()
		if err != nil {
			_, ok1 := err.(*db.ErrKeyExistsInBucket)
			_, ok2 := err.(*errors.ErrInsufficientData)
			if ok1 || ok2 {
				writeErrResp(w, r, http.StatusBadRequest, err)
			} else {
				writeErrResp(w, r, http.StatusInternalServerError, err)
			}
			return
		}
		if user.Roles.NumberOfElements() == 0 {
			user.Roles.Add(users.StandardUser)
		}
		messageBox := messages.NewMessageBox()
		user.MessageBox = messageBox.ID
		elementsToCreate = append(elementsToCreate, messageBox, user)
	}
	if err := db.Update(requestUser.UserName, elementsToCreate...); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: "users created successfully"})
}

// delete the user with the given name
func handleDelUser(w http.ResponseWriter, r *http.Request) {
	requestedUserName := mux.Vars(r)[userName]
	authenticatedUser := r.Context().Value(authenticatedUser).(*users.User)
	if requestedUserName == authenticatedUser.UserName {
		writeStrErrResp(w, r, http.StatusForbidden, "self deletion is forbidden")
		return
	}
	requestedUser, err := users.Get(requestedUserName)
	if err != nil {
		if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
			writeErrResp(w, r, http.StatusNotFound, err)
		} else {
			writeErrResp(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	if requestedUser.Roles.Contains(users.Admin) && !authenticatedUser.Roles.Contains(users.Admin) {
		writeStrErrResp(w, r, http.StatusForbidden, "deletion of admin user is forbidden")
		return
	}
	if err := users.Delete(requestedUser, fs.GetClient() != nil); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusOK, &Response{Message: fmt.Sprintf("user \"%s\" deleted successfully", requestedUser.UserName)})
}

// update the user with the given name with the given information
func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	requestedUserName := mux.Vars(r)[userName]
	exists, err := db.KeyExistsInBucket([]byte(db.Users), []byte(requestedUserName))
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	if !exists {
		writeStrErrResp(w, r, http.StatusNotFound, fmt.Sprintf("user named %s doesn't exist", requestedUserName))
		return
	}
	updatedUser := &users.User{}
	if err := json.NewDecoder(r.Body).Decode(updatedUser); err != nil {
		writeErrResp(w, r, http.StatusBadRequest, err)
		return
	}
	updatedUser.UserName = requestedUserName
	preUpdateUser, err := users.Get(requestedUserName)
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	updatedUser.Password = preUpdateUser.Password
	updatedUser.MessageBox = preUpdateUser.MessageBox
	updatedUser.CreatedOn = preUpdateUser.CreatedOn
	updatedUser.CreatedBy = preUpdateUser.CreatedBy
	if err := db.Update(r.Context().Value(authenticatedUser).(*users.User).UserName, updatedUser); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusAccepted, &Response{Message: fmt.Sprintf("user \"%s\" updated successfully", requestedUserName)})
}

// configure the users router
func initUsersRouter(r *mux.Router, manager *authManager) {
	usersBasePath := fmt.Sprintf("/%s", db.Users)
	usersRouter := r.PathPrefix(usersBasePath).Subrouter()
	usersRouter.HandleFunc("/", handleGetAllUsers).Methods(http.MethodGet)
	usersRouter.HandleFunc("/", handleRegisterUsers).Methods(http.MethodPost)
	manager.addPathToMap(fmt.Sprintf("%s/", usersBasePath), func (user *users.User, _ *http.Request) bool {
		return user.Roles.Contains(users.Secretary) || user.Roles.Contains(users.Admin)
	})
	specificUserPath := fmt.Sprintf("/{%s}", userName)
	usersRouter.HandleFunc(specificUserPath, handleGetUser).Methods(http.MethodGet)
	usersRouter.HandleFunc(specificUserPath, handleDelUser).Methods(http.MethodDelete)
	usersRouter.HandleFunc(specificUserPath, handleUpdateUser).Methods(http.MethodPut)
	manager.addRegex(regexp.MustCompile(fmt.Sprintf("%s/.", usersBasePath)), func (user *users.User, r *http.Request) bool {
		isSelfRequest := user.UserName == r.URL.Path[strings.LastIndex(r.URL.Path, "/") + 1 : ] // if the user is accessing his own user data
		return isSelfRequest || user.Roles.Contains(users.Secretary) || user.Roles.Contains(users.Admin)
	})
}
