package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/baidu/openedge/master"
	"github.com/baidu/openedge/module"
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

	m, err := master.New(f.Config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create master:", errors.Details(err))
		return
	}
	defer m.Close()
	err = m.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to start master:", errors.Details(err))
		return
	}

	module.Wait()
}
