package server

import (
	"github.com/DAv10195/submit_server/elements/users"
	"net/http"
	"regexp"
)

type authorizationFunc func(*users.User) bool

type regexpHandler struct {
	regexp	*regexp.Regexp
	invoke 	authorizationFunc
}

type regexpHandlerRegistry struct {
	regExpHandlers	[]*regexpHandler
}

type authManager struct {
	authMap map[string]authorizationFunc
	regexpHandlerRegistry
}

func (a *authManager) addPathToMap(path string, authFunc authorizationFunc) {
	a.authMap[path] = authFunc
}

func (a *authManager) addRegex(regex *regexp.Regexp, authFunc authorizationFunc) {
	a.regexpHandlerRegistry.regExpHandlers = append(a.regexpHandlerRegistry.regExpHandlers, &regexpHandler {
		regexp: regex,
		invoke: authFunc,
	})
}


func (a *authManager) authorizationMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		user := r.Context().Value(authenticatedUser).(*users.User)
		// try to get the handler with the path
		handler := a.authMap[path]
		// if path is  in the map. check the function.
		if handler != nil && !handler(user) {
			writeStrErrResp(w, r, http.StatusForbidden, accessDenied)
			return
		}
		for _, rHandler := range a.regExpHandlers {
			if rHandler.regexp.MatchString(path) && !rHandler.invoke(user) {
				writeStrErrResp(w, r, http.StatusForbidden, accessDenied)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
