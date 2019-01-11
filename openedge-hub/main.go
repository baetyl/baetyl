package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/module/logger"
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
	// 		fmt.Fprintln(os.Stderr, "Start profile failed: ", err.Error())
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

	f, err := module.ParseFlags("")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to parse argument:", err.Error())
		return
	}
	if f.Help {
		flag.Usage()
		return
	}

	m, err := New(f.Config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create module:", err.Error())
		logger.Log.WithError(err).Errorf("failed to create module")
		return
	}
	defer m.Close()
	err = m.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to start module:", err.Error())
		logger.Log.WithError(err).Errorf("failed to start module")
		return
	}
	module.Wait()
}
