package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/module"
	"github.com/juju/errors"
	// "net/http"
	// _ "net/http/pprof"
	// "path/filepath"
	// "runtime/trace"
)

func main() {

	// // go tool pprof http://localhost:6060/debug/pprof/profile
	// go func() {
	// 	err := http.ListenAndServe("localhost:6060", nil)
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, "Start profile failed: ", errors.Details(err))
	// 		return
	// 	}
	// }()

	// f, err := os.Create("trace.out")
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()

	// err = trace.Start(f)
	// if err != nil {
	// 	panic(err)
	// }
	// defer trace.Stop()

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
