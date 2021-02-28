package server

import (
	"context"
	"net/http"
	"sync"
)

type job func() error

type handlerWithWaitGroup func(w http.ResponseWriter, r *http.Request, wg *sync.WaitGroup)

// start a worker with the given id which listens for the cancellation of the given context and consumes the given
// job channel
func startWorker(ctx context.Context, wg *sync.WaitGroup, jobChan <- chan job , id int) {
	wg.Add(1)
	go func() {
		workerLogger := logger.WithField("worker", id)
		for {
			select {
				case <- ctx.Done():
					wg.Done()
					return
				case j := <- jobChan:
					if err := j(); err != nil {
						workerLogger.WithError(err).Error("error executing job")
					}
			}
		}
	}()
}

// a template for handling an incoming http request by a worker consuming the given job channel
func handleRequestByWorkers(handleFunc handlerWithWaitGroup, jobChan chan <- job) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			jobChan <- func() error {
				handleFunc(w, r, wg)
				return nil
			}
		}()
		wg.Wait()
	}
}
