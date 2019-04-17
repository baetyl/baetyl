package logger

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

var gLogger Logger

func init() {
	entry := logrus.NewEntry(logrus.New())
	entry.Level = logrus.InfoLevel
	entry.Logger.Out = ioutil.Discard
	entry.Logger.Level = logrus.InfoLevel
	entry.Logger.Formatter = newFormatter("text")
	gLogger = &logger{entry}
}

// Global returns the global logger
func Global() Logger {
	return gLogger
}

// WithField to global logger
func WithField(key string, value interface{}) Logger {
	return gLogger.WithField(key, value)
}

// WithError to global logger
func WithError(err error) Logger {
	return gLogger.WithError(err)
}

// Debugf to global logger
func Debugf(format string, args ...interface{}) {
	gLogger.Debugf(format, args...)
}

// Infof to global logger
func Infof(format string, args ...interface{}) {
	gLogger.Infof(format, args...)
}

// Warnf to global logger
func Warnf(format string, args ...interface{}) {
	gLogger.Warnf(format, args...)
}

// Errorf to global logger
func Errorf(format string, args ...interface{}) {
	gLogger.Errorf(format, args...)
}

// Fatalf to global logger
func Fatalf(format string, args ...interface{}) {
	gLogger.Fatalf(format, args...)
}

// Debugln to global logger
func Debugln(args ...interface{}) {
	gLogger.Debugln(args...)
}

// Infoln to global logger
func Infoln(args ...interface{}) {
	gLogger.Infoln(args...)
}

// Warnln to global logger
func Warnln(args ...interface{}) {
	gLogger.Warnln(args...)
}

// Errorln to global logger
func Errorln(args ...interface{}) {
	gLogger.Errorln(args...)
}

// Fatalln to global logger
func Fatalln(args ...interface{}) {
	gLogger.Fatalln(args)
}
