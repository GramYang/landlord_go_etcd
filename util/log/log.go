package log

import (
	"github.com/sirupsen/logrus"
)

var (
	L = getLog()
)

func getLog() *logrus.Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	l.SetLevel(logrus.TraceLevel)
	return l
}