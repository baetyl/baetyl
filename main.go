package main

import (
	_ "github.com/baetyl/baetyl/v2/ami/kube"
	_ "github.com/baetyl/baetyl/v2/ami/native"
	"github.com/baetyl/baetyl/v2/cmd"
	"github.com/baetyl/baetyl/v2/core"
	_ "github.com/baetyl/baetyl/v2/plugin/httplink"
	_ "github.com/baetyl/baetyl/v2/plugin/pubsub"
)

func init() {
	cmd.Hooks[cmd.HookNameNewCore] = core.NewCoreFunc(core.NewCore)
}

func main() {
	cmd.Execute()
}
