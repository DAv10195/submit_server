package server

import (
	"fmt"
	"github.com/DAv10195/submit_server/db"
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
		if r.URL.Path != fmt.Sprintf("/%s/%s", db.Users, register) {
			user, password, ok := r.BasicAuth()
			if !ok {
				http.Error(w, (&ErrorResponse{"no username/password given"}).String(), http.StatusUnauthorized)
				return
			}
			// authenticate the user associated with this request
			if err := users.Authenticate(user, password); err != nil {
				logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
				status := http.StatusInternalServerError
				if _, ok := err.(*users.ErrAuthenticationFailure); ok {
					status = http.StatusUnauthorized
				}
				http.Error(w, (&ErrorResponse{err.Error()}).String(), status)
				return
			}
			// write the name of the authenticated user on the request so it will be available to the eventual handler of
			// the request
			r.Header.Set(Authorization, user)
		}
		next.ServeHTTP(w, r)
	})
}

