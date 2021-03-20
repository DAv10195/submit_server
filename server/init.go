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
	am := NewAuthManager()
	baseRouter.Use(contentTypeMiddleware, authenticationMiddleware, am.authorizationMiddleware)
	agentsShutdown := initAgentsRouter(baseRouter, am)
	initUsersRouter(baseRouter, am)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      baseRouter,
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
	}
	server.RegisterOnShutdown(agentsShutdown)
	return server
}

func init() {
	agentMessageTypes.Add(MessageTypeKeepalive)
}
