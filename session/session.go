package session

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const (
	submitCookie		= "submit-server-cookie"
	sessionKeyFileName	= "submit_session.key"

	keyLength          = 32
	keyFilePerms       = 0600
	SubmitMaxCookieAge = 5 * 60
	SubmitSessionUser  = "submit_session_user"
)

var ErrNotFound = errors.New("session not found")
var ErrAlreadyExists = errors.New("session already exists")

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
