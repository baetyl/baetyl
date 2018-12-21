package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/baidu/openedge/master"
	"github.com/baidu/openedge/module"
)

func main() {
	f, err := module.ParseFlags(filepath.Join("etc", "openedge", "openedge.yml"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to parse argument:", err.Error())
		return
	}
	if f.Help {
		module.PrintUsage()
		return
	}

	m, err := master.New(f.WorkDir, f.Config)
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
