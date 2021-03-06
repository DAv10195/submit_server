package server

import "net/http"

type pathDetails struct {
	restrictedTo string //admin, user, stuff


}

type authManager struct {
	authMap map[string]pathDetails
}

func (a *authManager) authorizationMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.checkAccess(r) {
			//return status 403 - unauthorized
			status := http.StatusForbidden
			http.Error(w, (&ErrorResponse{"unauthorized user"}).String(), status)
			return
		}
		// serve the client
		next.ServeHTTP(w, r)
	})
}

func (a *authManager) checkAccess(r *http.Request)bool{
	req := r.RequestURI

}

