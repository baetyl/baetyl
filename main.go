package main

import (
	"github.com/baetyl/baetyl/cmd"
	_ "github.com/baetyl/baetyl/master/engine/docker"
	_ "github.com/baetyl/baetyl/master/engine/native"
)

func main() {
	cmd.Execute()
}
