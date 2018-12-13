package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/baidu/openedge/master"
	"github.com/baidu/openedge/module"
)

func main() {
	f, err := module.ParseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to parse argument:", err.Error())
		return
	}
	if f.Help {
		flag.Usage()
		return
	}

	m, err := master.New(f.Config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create master:", err.Error())
		return
	}
	defer m.Close()
	err = m.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to start master:", err.Error())
		return
	}

	module.Wait()
}
