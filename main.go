package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master"
	_ "github.com/baidu/openedge/master/engine/docker"
	_ "github.com/baidu/openedge/master/engine/native"
)

// compile variables
var (
	Version   string
	BuildTime string
	GoVersion string
)

const defaultConfig = "etc/openedge/openedge.yml"

func main() {
	exe, err := os.Executable()
	if err != nil {
		logger.Fatalln("failed to get executable path:", err.Error())
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		logger.Fatalln("failed to get realpath of executable:", err.Error())
	}
	workdir := path.Dir(path.Dir(exe))
	var flagW = flag.String("w", workdir, "working directory")
	var flagC = flag.String("c", defaultConfig, "config file path")
	var flagH = flag.Bool("h", false, "show this help")
	flag.Parse()
	if *flagH {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"OpenEdge version %s\nbuild time %s\n%s\n\n",
			Version,
			BuildTime,
			GoVersion,
		)
		flag.Usage()
		return
	}
	workdir, err = filepath.Abs(*flagW)
	if err != nil {
		logger.Fatalln("failed to get absolute path of workdir:", err.Error())
	}
	err = os.Chdir(workdir)
	if err != nil {
		logger.Fatalln("failed to change directory to workdir:", err.Error())
	}
	m, err := master.New(workdir, *flagC)
	if err != nil {
		logger.Fatalln("failed to create master:", err.Error())
	}
	defer m.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	<-sig
}
