package server

import (
	"github.com/DAv10195/submit_server/elements/users"
	"net/http"
	"regexp"
)

type authorizationFunc func(*users.User, *http.Request) bool

// an authorization handler using a regex
type regexpHandler struct {
	regexp	*regexp.Regexp
	invoke 	authorizationFunc
}

// authorization manager
type authManager struct {
	authMap 		map[string]authorizationFunc
	regExpHandlers	[]*regexpHandler
}

// return a new auth manager
func NewAuthManager() *authManager{
	am := &authManager{}
	am.authMap = make(map[string]authorizationFunc)
	return am
}

// middleware for invoking authorization handlers
func (a *authManager) authorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		user := r.Context().Value(authenticatedUser).(*users.User)
		// try to get the handler with the path
		handler := a.authMap[path]
		// if path is in the map invoke the handler.
		if handler != nil {
			if handler(user, r) {
				next.ServeHTTP(w, r)
			} else {
				writeStrErrResp(w, r, http.StatusForbidden, accessDenied)
			}
			return
		}
		// well, no direct handler, lets try matching regular expressions to find a handler
		for _, regExpHandler := range a.regExpHandlers {
			if regExpHandler.regexp.MatchString(path) && !regExpHandler.invoke(user, r) {
				writeStrErrResp(w, r, http.StatusForbidden, accessDenied)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// add a concrete path auth handler
func (a *authManager) addPathToMap(path string, authFunc authorizationFunc){
	a.authMap[path] = authFunc
}

// add a regex auth handler
func (a *authManager) addRegex(regex *regexp.Regexp, authFunc authorizationFunc) {
	a.regExpHandlers = append(a.regExpHandlers, &regexpHandler{
		regex,
		authFunc,
	})
}
