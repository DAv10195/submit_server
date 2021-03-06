package server

import "net/http"

type pathDetails struct {
	authFunc func(r *http.Request, w http.ResponseWriter)bool
}

type authManager struct {
	authMap map[string]*pathDetails
}

func (a *authManager) authorizationMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		uri = a.checkRegex(uri)
		if !a.authMap[uri].authFunc(r,w) {
			//return status 403 - unauthorized
			status := http.StatusForbidden
			http.Error(w, (&ErrorResponse{"unauthorized user"}).String(), status)
			return
		}
		// serve the client
		next.ServeHTTP(w, r)
	})
}

func (a * authManager) checkRegex(uri string) string{
	// analyze the path and return the string of the request group.
	return uri
}