package baetyl

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
)

// Run service
func Run(handle func(Context) error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "service is stopped with panic: %s\n%s", r, string(debug.Stack()))
		}
	}()

	signal.Ignore(syscall.SIGPIPE)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	c, err := newContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create context: %s\n", err.Error())
		return
	}
	err = c.setupMaster()
	if err != nil {
		c.log.Warnf("dial master fail: %s", err.Error())
		return
	}
	c.log.Infoln("service starting: ", os.Args)
	go func() {
		<-sig
		c.cancel()
	}()
	err = handle(c)
	if err != nil {
		c.log.WithError(err).Errorln("service is stopped with error")
	} else {
		c.log.Infoln("service stopped")
	}
}
