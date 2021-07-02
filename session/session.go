package session

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const (
	submitCookie		= "submit-server-cookie"
	sessionKeyFileName	= "submit_session.key"

	keyLength          			= 32
	keyFilePerms       			= 0600
	SubmitMaxCookieAge 			= 5 * 60
	SubmitSessionUser  			= "submit_session_user"

	authenticatedUser			= "authenticated_user"
)

var ErrNotFound = errors.New("session not found")
var ErrAlreadyExists = errors.New("session already exists")

type LoginData struct {
	UserName 		string		`json:"user_name"`
	Roles 			[]string	`json:"roles"`
	StaffCourses	[]string	`json:"staff_courses"`
	StudentCourses	[]string	`json:"student_courses"`
}

var store *sessions.CookieStore

func Init(dir string) error {
	key := make([]byte, keyLength)
	keyFileName := filepath.Join(dir, sessionKeyFileName)
	if _, err := os.Stat(keyFileName); err != nil {
		if os.IsNotExist(err) {
			if _, err := rand.Read(key); err != nil {
				return err
			}
			encodedKey := make([]byte, base64.StdEncoding.EncodedLen(keyLength))
			base64.StdEncoding.Encode(encodedKey, key)
			if err := ioutil.WriteFile(keyFileName, encodedKey, keyFilePerms); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		keyFromFile, err := ioutil.ReadFile(keyFileName)
		if err != nil {
			return err
		}
		decodedKey := make([]byte, keyLength)
		numDecodedBytes, err := base64.StdEncoding.Decode(decodedKey, keyFromFile)
		if numDecodedBytes != keyLength {
			return fmt.Errorf("number of bytes in key file (%s) is not as expected (%d)", keyFileName, keyLength)
		}
	}
	store = sessions.NewCookieStore(key)
	return nil
}

func Get(r *http.Request) (*sessions.Session, error) {
	sess, err := store.Get(r, submitCookie)
	if err != nil {
		return nil, err
	}
	if sess.IsNew {
		// delete temporary
		sess.Options.MaxAge = -1
		return nil, ErrNotFound
	}
	return sess, nil
}

func New(r *http.Request, userName string) (*sessions.Session, error) {
	sess, err := store.Get(r, submitCookie)
	if err != nil {
		return nil, err
	}
	if !sess.IsNew {
		return nil, ErrAlreadyExists
	}
	sess.Values[SubmitSessionUser] = userName
	sess.Options.MaxAge = SubmitMaxCookieAge
	return sess, nil
}

func writeErr(w http.ResponseWriter, errStr string, logger *logrus.Entry) {
	w.WriteHeader(http.StatusInternalServerError)
	type resp struct {
		Message string `json:"message"`
	}
	res := &resp{errStr}
	respBytes, _ := json.Marshal(res)
	if _, err := w.Write(respBytes); err != nil && logger != nil {
		logger.WithError(err).Error("error writing login error")
	}
}

func LoginHandler(logger *logrus.Entry) func (w http.ResponseWriter, r *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(authenticatedUser).(*users.User)
		if !ok {
			errMsg := "no user in request"
			if logger != nil {
				logger.Error(errMsg)
			}
			writeErr(w, errMsg, logger)
			return
		}
		ld := &LoginData{UserName: user.UserName, Roles: user.Roles.Slice(), StaffCourses: user.CoursesAsStaff.Slice(), StudentCourses: user.CoursesAsStudent.Slice()}
		ldBytes, err := json.Marshal(ld)
		if err != nil {
			errMsg := "error formatting login data"
			if logger != nil {
				logger.WithError(err).Error(errMsg)
			}
			writeErr(w, err.Error(), logger)
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(ldBytes); err != nil && logger != nil {
			logger.WithError(err).Error("error writing login data")
		}
	}
}

func InitSessionForTest() func() {
	dir := os.TempDir()
	if err := Init(dir); err != nil {
		panic(err)
	}
	return func() {
		if err := os.RemoveAll(filepath.Join(dir, sessionKeyFileName)); err != nil {
			panic(err)
		}
	}
}
