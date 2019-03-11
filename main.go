package main

import (
	"github.com/baidu/openedge/cmd"
	_ "github.com/baidu/openedge/master/engine/docker"
	_ "github.com/baidu/openedge/master/engine/native"
)

// compile variables
var (
	Version   string
	BuildTime string
	GoVersion string
)

func main() {
	cmd.Execute()
}
