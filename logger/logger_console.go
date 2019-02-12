package logger

import (
	"fmt"
	"os"
)

type stdLogger struct{}

func (l *stdLogger) WithField(key string, value interface{}) Logger {
	return l
}

func (l *stdLogger) WithError(err error) Logger {
	return l
}

func (l *stdLogger) Debugf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

func (l *stdLogger) Infof(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

func (l *stdLogger) Warnf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (l *stdLogger) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (l *stdLogger) Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func (l *stdLogger) Debugln(args ...interface{}) {
	fmt.Fprintln(os.Stdout, args...)
}

func (l *stdLogger) Infoln(args ...interface{}) {
	fmt.Fprintln(os.Stdout, args...)
}

func (l *stdLogger) Warnln(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

func (l *stdLogger) Errorln(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

func (l *stdLogger) Fatalln(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}
