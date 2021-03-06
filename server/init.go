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
	authManager.authMap[GetUser] = &pathDetails{
		authFunc: func(r *http.Request, w http.ResponseWriter) bool {
			requestUserName := mux.Vars(r)[userName]
			user, err := users.Get(requestUserName)
			if err != nil {
				logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
				status := http.StatusInternalServerError
				if _, ok := err.(*db.ErrKeyNotFoundInBucket); ok {
					status = http.StatusNotFound
				}
				writeErrResp(w,r,status,err)
				return false
			}
			user = r.Context().Value(authenticatedUser).(*users.User)
			if requestUserName != user.UserName && !user.Roles.Contains(users.Secretary) && !user.Roles.Contains(users.Admin) {
				logger.Errorf("access to \"%s\" denied for user \"%s\"", r.URL.Path, user.UserName)
				writeStrErrResp(w,r,http.StatusInternalServerError,"access forbidden")
				return false
			}
			return true
		},
	}
	authManager.authMap[GetAllUsers] = &pathDetails{
		authFunc: func(r *http.Request, w http.ResponseWriter) bool {
			requestUserName := r.Context().Value(authenticatedUser).(*users.User).UserName
			user, err := users.Get(requestUserName)
			if err != nil {
				logger.WithError(err).Errorf(logHttpErrFormat, r.URL.Path)
				writeErrResp(w,r,http.StatusInternalServerError,err)
				return false
			}
			if !user.Roles.Contains(users.Secretary) && !users.Roles.Contains(users.Admin) {
				user := r.Context().Value(authenticatedUser).(*users.User)
				logger.Errorf("access to \"%s\" denied for user \"%s\"", r.URL.Path, user.UserName)
				writeStrErrResp(w,r,http.StatusForbidden,"access forbidden")
				return false
			}
			return true
		},
	}
	initUsersRouter(baseRouter)
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      baseRouter,
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
	}
}
