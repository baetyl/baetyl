package openedge

import (
	"fmt"
	"os"
)

type fakeLogger struct{}

func (l *fakeLogger) WithField(key string, value interface{}) Logger {
	return l
}

func (l *fakeLogger) WithError(err error) Logger {
	return l
}

func (l *fakeLogger) Debugf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (l *fakeLogger) Infof(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (l *fakeLogger) Warnf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (l *fakeLogger) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (l *fakeLogger) Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func (l *fakeLogger) Debugln(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

func (l *fakeLogger) Infoln(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

func (l *fakeLogger) Warnln(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

func (l *fakeLogger) Errorln(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

func (l *fakeLogger) Fatalln(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}
