package server

import (
	"context"
	"github.com/DAv10195/submit_server/elements/users"
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
		var userStruct *users.User
		user, password, ok := r.BasicAuth()
		if !ok {
			writeStrErrResp(w, r, http.StatusUnauthorized, "no username/password given")
			return
		}
		// authenticate the user associated with this request
		var err error
		userStruct, err = users.Authenticate(user, password)
		if err != nil {
			if _, ok := err.(*users.ErrAuthenticationFailure); ok {
				writeErrResp(w, r, http.StatusUnauthorized, err)
			} else {
				writeErrResp(w, r, http.StatusInternalServerError, err)
			}
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authenticatedUser, userStruct)))
	})
}
