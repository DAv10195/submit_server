package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/DAv10195/submit_server/session"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
)

func InitServer(port int, tlsConf *tls.Config, wg *sync.WaitGroup, ctx context.Context) *http.Server {
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
	initMossRequestRouter(baseRouter, am)
	initMessagesRouter(baseRouter, am)
	initFilesRouter(baseRouter, am)
	initAgentsBackend(baseRouter, am, ctx, wg)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      baseRouter,
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
		TLSConfig:	  tlsConf,
	}
	server.RegisterOnShutdown(func () {
		defer wg.Done()
		agentEndpoints.close()
	})
	return server
}

func GetTlsConfig(certFilePath, keyFilePath string) (*tls.Config, error) {
	if certFilePath != "" && keyFilePath != "" {
		cert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
		if err != nil {
			return nil, err
		}
		return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
	}
	return nil, nil
}
