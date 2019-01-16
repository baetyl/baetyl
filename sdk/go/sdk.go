package sdk

import (
	"os"

	openedge "github.com/baidu/openedge/api/go"
)

// Run service
func Run(handler func(openedge.Context) error) {
	ctx, err := newContext()
	if err != nil {
		openedge.Fatalln("failed to create context:", err.Error())
	}
	defer ctx.Close()
	err = handler(ctx)
	if err != nil {
		openedge.Errorln("failed to run service:", err.Error())
		os.Exit(1)
	}
}
