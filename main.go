package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	openedge "github.com/baidu/openedge/api/go"
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
		openedge.Fatalln("get executable path fail:", err.Error())
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		openedge.Fatalln("get realpath of executable fail:", err.Error())
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
		openedge.Fatalln("get absolute path of workdir fail:", err.Error())
	}
	err = os.Chdir(workdir)
	if err != nil {
		openedge.Fatalln("change dir to workdir fail:", err.Error())
	}

	m, err := master.New(workdir, *flagC)
	if err != nil {
		openedge.Fatalln("failed to create master:", err.Error())
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	<-sig
	m.Close()
}
