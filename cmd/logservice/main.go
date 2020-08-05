package main

import (
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
)

func main() {
	context.Run(func(ctx context.Context) error {
		for {
			select {
			case <-time.After(time.Second):
				ctx.Log().Info("log a message")
			case <-ctx.WaitChan():
				return nil
			}
		}
	})
}
