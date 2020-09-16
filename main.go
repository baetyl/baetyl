package main

import (
	_ "github.com/baetyl/baetyl/ami/kube"
	_ "github.com/baetyl/baetyl/ami/native"
	"github.com/baetyl/baetyl/cmd"
	_ "github.com/baetyl/baetyl/plugin/httplink"
	_ "github.com/baetyl/baetyl/plugin/memory"
)

func main() {
	cmd.Execute()
}
