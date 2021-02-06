package server

import (
	"context"
	"github.com/gorilla/mux"
)

func initServer(ctx context.Context) {
	
	baseRouter := mux.NewRouter()
	baseRouter.Use(contentTypeMiddleware)
	baseRouter.Use(authenticationMiddleware)


}
