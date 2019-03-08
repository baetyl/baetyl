package daemon

import (
	"errors"
	"os"
	"syscall"
)

// ErrStop errStop
var ErrStop = errors.New("stop serve signals")

// SignalHandlerFunc signal handler
type SignalHandlerFunc func(sig os.Signal) (err error)

var handlers = make(map[os.Signal]SignalHandlerFunc)

func init() {
	handlers[syscall.SIGTERM] = sigtermHandler
}

func sigtermHandler(sig os.Signal) error {
	return ErrStop
}
