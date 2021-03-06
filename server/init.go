package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func InitServer(cfg *Config) *http.Server {
	logger.Info("initializing server...")
	// configure base request router and content type middleware
	baseRouter := mux.NewRouter()
	baseRouter.Use(contentTypeMiddleware)
	baseRouter.Use(authenticationMiddleware)
	am := &authManager{authMap: make(map[string]authorizationFunc)}
	baseRouter.Use(am.authorizationMiddleware)
	initUsersRouter(baseRouter, am)
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      baseRouter,
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
	}
}
