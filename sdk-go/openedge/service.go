package openedge

import (
	"github.com/baidu/openedge/logger"
)

// Run service
func Run(handle func(Context) error) {
	ctx, err := newContext()
	if err != nil {
		logger.Fatalln("failed to create context:", err.Error())
	}
	defer ctx.Close()
	err = handle(ctx)
	if err != nil {
		logger.Fatalln("failed to run service:", err.Error())
	}
}
