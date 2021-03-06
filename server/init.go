package server

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
)

func InitServer(ctx context.Context, cfg *Config) (*http.Server, *sync.WaitGroup) {
	wg := &sync.WaitGroup{}
	// start workers to handle jobs generated by incoming requests to the rest api
	jobChan := make(chan job, cfg.NumberOfServerGoroutines)
	logger.Info("starting server workers...")
	for i := 1; i <= cfg.NumberOfServerGoroutines; i++ {
		startWorker(ctx, wg, jobChan, i)
	}
	authManager := authManager{}

	// configure base request router and content type middleware
	baseRouter := mux.NewRouter()
	baseRouter.Use(contentTypeMiddleware)
	baseRouter.Use(authenticationMiddleware)
	baseRouter.Use(authManager.authorizationMiddleware)
	initUsersRouter(baseRouter, jobChan)
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      baseRouter,
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
	}, wg
}
