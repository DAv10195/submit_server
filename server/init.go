package server

import (
	"context"
	"fmt"
	"github.com/DAv10195/submit_server/session"
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
	baseRouter.HandleFunc("/", session.LoginHandler(logger)).Methods(http.MethodGet)
	initUsersRouter(baseRouter, am)
	initCoursesRouter(baseRouter, am)
	initAssDefsRouter(baseRouter, am)
	initAssInstsRouter(baseRouter, am)
	initAppealsRouter(baseRouter, am)
	initTestsRouter(baseRouter, am)
	initTestRequestsRouter(baseRouter, am)
	initMessagesRouter(baseRouter, am)
	initFilesRouter(baseRouter, am)
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
