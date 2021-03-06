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
	logger.Debugf("handling request: %v", r)
	requestUserName := mux.Vars(r)[userName]
	user, err := users.Get(requestUserName)
	userBytes, err := json.Marshal(user)
	if err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
		http.Error(w, (&ErrorResponse{err.Error()}).String(), http.StatusInternalServerError)
		return
	}
	if _, err = w.Write(userBytes); err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
	}
}

func handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("handling request: %v", r)
	var resp struct {
		Users []*users.User `json:"users"`
	}
	if err := db.QueryBucket([]byte(db.Users), func(_, elementBytes []byte) error {
		user := &users.User{}
		if err := json.Unmarshal(elementBytes, user); err != nil {
			return err
		}
		resp.Users = append(resp.Users, user)
		return nil
	}); err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
		http.Error(w, (&ErrorResponse{err.Error()}).String(), http.StatusInternalServerError)
		return
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
		http.Error(w, (&ErrorResponse{err.Error()}).String(), http.StatusInternalServerError)
		return
	}
	if _, err = w.Write(respBytes); err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
	}
}

func handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	user := &users.User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
		http.Error(w, (&ErrorResponse{err.Error()}).String(), http.StatusInternalServerError)
		return
	}
	if user.UserName == "" {
		err := fmt.Errorf("missing user name")
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
		http.Error(w, (&ErrorResponse{err.Error()}).String(), http.StatusBadRequest)
		return
	}
	exists, err := db.KeyExistsInBucket([]byte(db.Users), []byte(user.UserName))
	if err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
		http.Error(w, (&ErrorResponse{err.Error()}).String(), http.StatusInternalServerError)
		return
	}
	if exists {
		logger.WithError(&db.ErrKeyExistsInBucket{Bucket: db.Users, Key: user.UserName}).Errorf(logHttpErrFormat, r.URL.Path)
		http.Error(w, (&ErrorResponse{err.Error()}).String(), http.StatusBadRequest)
		return
	}
	if user.Password == "" {
		err := fmt.Errorf("missing password")
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
		http.Error(w, (&ErrorResponse{err.Error()}).String(), http.StatusBadRequest)
		return
	}
	if err = users.ValidateEmail(user.Email); err != nil {
		err := fmt.Errorf("missing password")
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
		http.Error(w, (&ErrorResponse{err.Error()}).String(), http.StatusBadRequest)
		return
	}
	encryptedPassword, err := db.Encrypt(user.Password)
	if err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
		http.Error(w, (&ErrorResponse{err.Error()}).String(), http.StatusInternalServerError)
		return
	}
	user.Password = encryptedPassword
	user.Roles = containers.NewStringSet()
	user.Roles.Add(users.StandardUser)
	user.CoursesAsStaff = containers.NewStringSet()
	user.CoursesAsStudent = containers.NewStringSet()
	messageBox := messages.NewMessageBox()
	if err := db.Update(db.System, messageBox, user); err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
		http.Error(w, (&ErrorResponse{err.Error()}).String(), http.StatusInternalServerError)
		return
	}
	var resp struct {
		Message	string	`json:"message"`
	}
	resp.Message = fmt.Sprintf("user \"%s\" created successfully", user.UserName)
	respBytes, err := json.Marshal(resp)
	if err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
		http.Error(w, (&ErrorResponse{err.Error()}).String(), http.StatusInternalServerError)
		return
	}
	if _, err = w.Write(respBytes); err != nil {
		logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
	}
}

// configure the users router
func initUsersRouter(r *mux.Router) {
	usersRouter := r.PathPrefix(fmt.Sprintf("/%s", db.Users)).Subrouter()
	usersRouter.HandleFunc("/", handleGetAllUsers).Methods(http.MethodGet)
	usersRouter.HandleFunc(fmt.Sprintf("/%s", register), handleRegisterUser).Methods(http.MethodPost)
	usersRouter.HandleFunc(fmt.Sprintf("/{%s}", userName), handleGetUser).Methods(http.MethodGet)
}
