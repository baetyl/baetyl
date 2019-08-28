package baetyl

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/baetyl/baetyl/logger"
)

// Run service
func Run(handle func(Context) error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "service is stopped with panic: %s\n%s", r, string(debug.Stack()))
		}
	}()
	c, err := newContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s][%s] failed to create context: %s\n", c.sn, c.in, err.Error())
		logger.WithError(err).Errorln("failed to create context")
		return
	}
	logger.Infoln("service starting: ", os.Args)
	err = handle(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s][%s] service is stopped with error: %s\n", c.sn, c.in, err.Error())
		logger.WithError(err).Errorln("service is stopped with error")
	} else {
		logger.Infoln("service stopped")
	}
}
