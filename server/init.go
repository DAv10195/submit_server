package server

import (
	"fmt"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/gorilla/mux"
	"net/http"
)


func InitServer(cfg *Config) *http.Server {
	logger.Info("initializing server...")
	// configure base request router and content type middleware
	baseRouter := mux.NewRouter()
	baseRouter.Use(contentTypeMiddleware)
	baseRouter.Use(authenticationMiddleware)
	authManager := authManager{}
	authManager.authMap = make(map[string]*pathDetails)
	authManager.authMap[users.GetUser] = &pathDetails{
		authFunc: func(r *http.Request, w http.ResponseWriter) bool {
			requestUserName := mux.Vars(r)[userName]
			user, err := users.Get(requestUserName)
			if err != nil {
				logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
				status := http.StatusInternalServerError
				if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
					status = http.StatusNotFound
				}
				http.Error(w, (&ErrorResponse{err.Error()}).String(), status)
				return false
			}
			if requestUserName != r.Header.Get(Authorization) && !user.Roles.Contains(users.Secretary) && !user.Roles.Contains(users.Admin) {
				logger.Errorf("access to \"%s\" denied for user \"%s\"", r.URL.Path, r.Header.Get(Authorization))
				http.Error(w, (&ErrorResponse{accessDenied}).String(), http.StatusForbidden)
				return false
			}
			return true
		},
	}
	authManager.authMap[users.GetAllUsers] = &pathDetails{
		authFunc: func(r *http.Request, w http.ResponseWriter) bool {
			requestUserName := r.Header.Get(Authorization)
			user, err := users.Get(requestUserName)
			if err != nil {
				logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
				http.Error(w, (&ErrorResponse{err.Error()}).String(), http.StatusInternalServerError)
				return false
			}
			if !user.Roles.Contains(users.Secretary) && !users.Roles.Contains(users.Admin) {
				logger.Errorf("access to \"%s\" denied for user \"%s\"", r.URL.Path, r.Header.Get(Authorization))
				http.Error(w, (&ErrorResponse{accessDenied}).String(), http.StatusForbidden)
				return false
			}
			return true
		},
	}
	baseRouter.Use(authManager.authorizationMiddleware)
	initUsersRouter(baseRouter)
	initUsersRouter(baseRouter)
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      baseRouter,
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
	}
}
