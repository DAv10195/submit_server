package server

import (
	"context"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/session"
	"net/http"
)

func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContentType, ApplicationJson)
		next.ServeHTTP(w, r)
	})
}

// authenticate incoming requests
func authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := session.Get(r)
		if err == nil {
			user, err := users.Get(sess.Values[session.SubmitSessionUser].(string))
			if err != nil {
				writeErrResp(w, r, http.StatusInternalServerError, err)
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authenticatedUser, user)))
			return
		} else if err != session.ErrNotFound {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
		var userStruct *users.User
		user, password, ok := r.BasicAuth()
		if !ok {
			writeStrErrResp(w, r, http.StatusUnauthorized, "no username/password given")
			return
		}
		// authenticate the user associated with this request
		userStruct, err = users.Authenticate(user, password)
		if err != nil {
			if _, ok := err.(*users.ErrAuthenticationFailure); ok {
				writeErrResp(w, r, http.StatusUnauthorized, err)
			} else {
				writeErrResp(w, r, http.StatusInternalServerError, err)
			}
			return
		}
		sess, err = session.New(r, user)
		if err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
		if err := sess.Save(r, w); err != nil {
			writeErrResp(w, r, http.StatusInternalServerError, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authenticatedUser, userStruct)))
	})
}
