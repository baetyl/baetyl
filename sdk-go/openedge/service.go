package openedge

import (
	"os"

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
		logger.Errorln("failed to run service:", err.Error())
		os.Exit(1)
	}
}
