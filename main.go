package main

import (
	_ "github.com/baetyl/baetyl/v2/ami/kube"
	_ "github.com/baetyl/baetyl/v2/ami/native"
	"github.com/baetyl/baetyl/v2/cmd"
	"github.com/baetyl/baetyl/v2/core"
	"github.com/baetyl/baetyl/v2/initz"
	_ "github.com/baetyl/baetyl/v2/plugin/httplink"
	_ "github.com/baetyl/baetyl/v2/plugin/pubsub"

	_ "github.com/baetyl/baetyl/v2/plugin/mqttlink"
	_ "github.com/baetyl/baetyl/v2/plugin/nodestats"
	_ "github.com/baetyl/baetyl/v2/plugin/nvstats"
	_ "github.com/baetyl/baetyl/v2/plugin/qpsstats"
	_ "github.com/baetyl/baetyl/v2/plugin/wslink"
)

func init() {
	cmd.Hooks[cmd.HookNameStartCoreService] = core.StartCoreServiceFunc(core.StartCoreService)
	cmd.Hooks[cmd.HookNameStartInitService] = initz.StartInitServiceFunc(initz.StartInitService)
	cmd.Hooks[cmd.HookGetFingerprint] = cmd.GetFingerprintFunc(cmd.GetFingerprint)
}

func main() {
	cmd.Execute()
}
