package db

import "github.com/sirupsen/logrus"

var logger *logrus.Entry

func init() {
	logger = logrus.WithFields(logrus.Fields{"component":"db"})
}
