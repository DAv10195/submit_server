package server

import (
	"fmt"
	submitws "github.com/DAv10195/submit_commons/websocket"
	"github.com/gorilla/mux"
	"net/http"
)

func InitServer(cfg *Config) *http.Server {
	logger.Info("initializing server...")
	// configure router and middleware
	baseRouter := mux.NewRouter()
	am := NewAuthManager()
	baseRouter.Use(contentTypeMiddleware, authenticationMiddleware, am.authorizationMiddleware)
	initUsersRouter(baseRouter, am)
	baseRouter.HandleFunc(submitws.Agents, agentEndpoints.agentsEndpoint)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      baseRouter,
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
	}
	server.RegisterOnShutdown(agentEndpoints.close)
	return server
}
