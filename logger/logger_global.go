package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Global the global logger
var Global Logger

func init() {
	entry := logrus.NewEntry(logrus.New())
	entry.Level = logrus.InfoLevel
	entry.Logger.Out = os.Stdout
	entry.Logger.Level = logrus.InfoLevel
	entry.Logger.Formatter = newFormatter("text")
	Global = &logger{entry}
}

// WithField to global logger
func WithField(key string, value interface{}) Logger {
	return Global.WithField(key, value)
}

// WithError to global logger
func WithError(err error) Logger {
	return Global.WithError(err)
}

// Debugf to global logger
func Debugf(format string, args ...interface{}) {
	Global.Debugf(format, args...)
}

// Infof to global logger
func Infof(format string, args ...interface{}) {
	Global.Infof(format, args...)
}

// Warnf to global logger
func Warnf(format string, args ...interface{}) {
	Global.Warnf(format, args...)
}

// Errorf to global logger
func Errorf(format string, args ...interface{}) {
	Global.Errorf(format, args...)
}

// Fatalf to global logger
func Fatalf(format string, args ...interface{}) {
	Global.Fatalf(format, args...)
}

// Debugln to global logger
func Debugln(args ...interface{}) {
	Global.Debugln(args...)
}

// Infoln to global logger
func Infoln(args ...interface{}) {
	Global.Infoln(args...)
}

// Warnln to global logger
func Warnln(args ...interface{}) {
	Global.Warnln(args...)
}

// Errorln to global logger
func Errorln(args ...interface{}) {
	Global.Errorln(args...)
}

// Fatalln to global logger
func Fatalln(args ...interface{}) {
	Global.Fatalln(args)
}
