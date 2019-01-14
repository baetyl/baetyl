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
	_ "github.com/baidu/openedge/master/engine/native"
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
		fmt.Fprintf(flag.CommandLine.Output(), "Version of %s: %s\n", os.Args[0], master.Version)
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
