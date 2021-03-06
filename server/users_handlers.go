package server

import (
	"encoding/json"
	"fmt"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/messages"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/util/containers"
	"github.com/gorilla/mux"
	"net/http"
)

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(authenticatedUser).(*users.User)
	requestedUserName := mux.Vars(r)[userName]
	isSelfRequest := requestedUserName == user.UserName
	if !isSelfRequest && !user.Roles.Contains(users.Secretary) && !user.Roles.Contains(users.Admin) {
		writeStrErrResp(w, r, http.StatusForbidden, accessDenied)
		return
	}
	var requestedUser *users.User
	if isSelfRequest {
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
	user := r.Context().Value(authenticatedUser).(*users.User)
	if !user.Roles.Contains(users.Secretary) && !users.Roles.Contains(users.Admin) {
		writeStrErrResp(w, r, http.StatusForbidden, accessDenied)
		return
	}
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
	var body struct {
		Users	[]*users.User	`json:"users"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	var elementsToCreate []db.IBucketElement
	for _, user := range body.Users {
		if err := users.ValidateNew(user); err != nil {
			_, ok1 := err.(*db.ErrKeyExistsInBucket)
			_, ok2 := err.(*users.ErrInsufficientData)
			if ok1 || ok2 {
				writeErrResp(w, r, http.StatusBadRequest, err)
			} else {
				writeErrResp(w, r, http.StatusInternalServerError, err)
			}
			return
		}
		encryptedPassword, err := db.Encrypt(user.Password)
		if err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
		user.Password = encryptedPassword
		user.Roles = containers.NewStringSet()
		user.Roles.Add(users.StandardUser)
		user.CoursesAsStaff = containers.NewStringSet()
		user.CoursesAsStudent = containers.NewStringSet()
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
	isSelfRequest := requestedUserName == user.UserName
	if !isSelfRequest && !user.Roles.Contains(users.Secretary) && !user.Roles.Contains(users.Admin) {
		writeStrErrResp(w, r, http.StatusForbidden, accessDenied)
		return
	}
	var requestedUser *users.User
	if isSelfRequest {
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
	if err := db.Delete(requestedUser); err != nil {
		writeErrResp(w, r, http.StatusInternalServerError, err)
		return
	}
	writeResponse(w, r, http.StatusOK, &Response{fmt.Sprintf("user \"%s\" deleted successfully", requestedUser.UserName)})
}

// configure the users router
func initUsersRouter(r *mux.Router) {
	usersRouter := r.PathPrefix(fmt.Sprintf("/%s", db.Users)).Subrouter()
	usersRouter.HandleFunc("/", handleGetAllUsers).Methods(http.MethodGet)
	usersRouter.HandleFunc("/", handleRegisterUsers).Methods(http.MethodPost)
	usersRouter.HandleFunc(fmt.Sprintf("/{%s}", userName), handleGetUser).Methods(http.MethodGet)
	usersRouter.HandleFunc(fmt.Sprintf("/{%s}", userName), handleDelUser).Methods(http.MethodDelete)
}
