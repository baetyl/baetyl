package openedge

import (
	"github.com/baidu/openedge/logger"
)

// Run service
func Run(handle func(Context) error) {
	c, err := newContext()
	if err != nil {
		logger.Fatalln("failed to create context:", err.Error())
		return
	}
	logger.Debugln("service starting")
	err = handle(c)
	if err != nil {
		logger.Fatalln("failed to run service:", err.Error())
	} else {
		logger.Debugln("service stopped")
	}
}
