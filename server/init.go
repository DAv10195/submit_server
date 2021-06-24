package server

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
)

func InitServer(port int, wg *sync.WaitGroup, ctx context.Context) *http.Server {
	logger.Info("initializing server...")
	// configure router and middleware
	baseRouter := mux.NewRouter()
	am := NewAuthManager()
	baseRouter.Use(contentTypeMiddleware, authenticationMiddleware, am.authorizationMiddleware)
	initUsersRouter(baseRouter, am)
	initCoursesRouter(baseRouter, am)
	initAssDefsRouter(baseRouter, am)
	initAgentsBackend(baseRouter, am, ctx, wg)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      baseRouter,
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
	}
	server.RegisterOnShutdown(func () {
		defer wg.Done()
		agentEndpoints.close()
	})
	return server
}
