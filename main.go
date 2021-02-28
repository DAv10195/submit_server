package main

import (
	"context"
	"github.com/DAv10195/submit_server/cmd"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logger := logrus.WithFields(logrus.Fields{"component":"main"})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer cancel()
		logger.Infof("signal received: %v", <-signalChan)
	}()
	rootCmd := cmd.NewRootCmd(ctx, os.Args)
	if err := rootCmd.Execute(); err != nil {
		logger.WithError(err).Fatal("error running submit server")
	}
}
