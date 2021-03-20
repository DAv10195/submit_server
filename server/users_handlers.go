package server

import (
	"encoding/json"
	"fmt"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/messages"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/util"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
	"strings"
)

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

func handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	var elements []db.IBucketElement
	if err := db.QueryBucket([]byte(db.Users), func(_, elementBytes []byte) error {
		user := &users.User{}
		if err := json.Unmarshal(elementBytes, user); err != nil {
			return err
		}
		elements = append(elements, user)
		return nil
	}); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeElements(w, r, http.StatusOK, elements)
}

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
			_, ok2 := err.(*util.ErrInsufficientData)
			if ok1 || ok2 {
				writeErrResp(w, r, http.StatusBadRequest, err)
			} else {
				writeErrResp(w, r, http.StatusInternalServerError, err)
			}
			return
		}
		messageBox := messages.NewMessageBox()
		user.MessageBox = messageBox.ID
		elementsToCreate = append(elementsToCreate, messageBox, user)
	}
	if err := db.Update(db.System, elementsToCreate...); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusOK, &Response{"users created successfully"})
}

func handleDelUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(authenticatedUser).(*users.User)
	requestedUserName := mux.Vars(r)[userName]
	if requestedUserName == user.UserName {
		writeStrErrResp(w, r, http.StatusBadRequest, "self deletion is forbidden")
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
	if err := db.Delete(requestedUser); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusOK, &Response{fmt.Sprintf("user \"%s\" deleted successfully", requestedUser.UserName)})
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(authenticatedUser).(*users.User)
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
	if requestedUserName != updatedUser.UserName {
		writeStrErrResp(w, r, http.StatusBadRequest, "updating user name is forbidden")
		return
	}
	preUpdateUser, err := users.Get(requestedUserName)
	if err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	if preUpdateUser.Password != updatedUser.Password {
		encryptedPassword, err := db.Encrypt(updatedUser.Password)
		if err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
		updatedUser.Password = encryptedPassword
	}
	if err := db.Update(user.UserName, updatedUser); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusOK, &Response{fmt.Sprintf("user \"%s\" updated successfully", requestedUserName)})
}

// configure the users router
func initUsersRouter(r *mux.Router, manager *authManager) {
	usersBasePath := fmt.Sprintf("/%s", db.Users)
	usersRouter := r.PathPrefix(usersBasePath).Subrouter()
	usersRouter.HandleFunc("/", handleGetAllUsers).Methods(http.MethodGet)
	usersRouter.HandleFunc("/", handleRegisterUsers).Methods(http.MethodPost)
	manager.addPathToMap(fmt.Sprintf("%s/", usersBasePath), func (user *users.User, _ string) bool {
		return user.Roles.Contains(users.Secretary) || user.Roles.Contains(users.Admin)
	})
	specificUserPath := fmt.Sprintf("/{%s}", userName)
	usersRouter.HandleFunc(specificUserPath, handleGetUser).Methods(http.MethodGet)
	usersRouter.HandleFunc(specificUserPath, handleDelUser).Methods(http.MethodDelete)
	usersRouter.HandleFunc(specificUserPath, handleUpdateUser).Methods(http.MethodPut)
	manager.addRegex(regexp.MustCompile(fmt.Sprintf("%s/.", usersBasePath)), func (user *users.User, path string) bool {
		isSelfRequest := user.UserName == path[strings.LastIndex(path, "/") + 1 : ] // if the user is accessing his own user data
		return isSelfRequest || user.Roles.Contains(users.Secretary) || user.Roles.Contains(users.Admin)
	})
}
