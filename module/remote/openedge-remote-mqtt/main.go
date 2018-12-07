package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/baidu/openedge/logger"
	module "github.com/baidu/openedge/module"
	"github.com/juju/errors"
)

func main() {
	f, err := module.ParseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse argument:", errors.Details(err))
		return
	}
	if f.Help {
		flag.Usage()
		return
	}

	m, err := New(f.Config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create module:", errors.Details(err))
		logger.WithError(err).Errorf("Failed to create module")
		return
	}
	defer m.Close()
	err = m.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to start module:", errors.Details(err))
		logger.WithError(err).Errorf("Failed to start module")
		return
	}
	module.Wait()
}
