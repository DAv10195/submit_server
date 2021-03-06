package server

import (
	"net/http"
	"regexp"
)

type pathDetails struct {
	authFunc func(r *http.Request, w http.ResponseWriter)bool
}

type authManager struct {
	authMap map[string]*pathDetails
}

func (a *authManager) authorizationMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		// try to get the handler with the uri
		uriHandler := a.authMap[uri]
		// if its found in the map.
		if uriHandler != nil && uriHandler.authFunc(r,w) {
			// if uri is  in the map. check the function.
			next.ServeHTTP(w, r)
			return
		}
		// if its not found in the map. try to find a handler using regex.
		pathType := a.checkPathType(uri)
		if pathType != ""  && a.authMap[pathType].authFunc(r,w) {
			next.ServeHTTP(w, r)
			return
		}
		//return status 403 - unauthorized if handler is not found.
		status := http.StatusForbidden
		http.Error(w, (&ErrorResponse{"unauthorized user"}).String(), status)
		return
	})
}

func (a * authManager) checkPathType(uri string) string{
	// analyze the path and return the string of the request group.
	matchUser, _ := regexp.MatchString("users/.", uri)
	if matchUser{
		return "/user"
	}
	return ""
}