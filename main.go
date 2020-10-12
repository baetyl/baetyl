package main

import (
	_ "github.com/baetyl/baetyl/v2/ami/kube"
	_ "github.com/baetyl/baetyl/v2/ami/native"
	"github.com/baetyl/baetyl/v2/cmd"
	_ "github.com/baetyl/baetyl/v2/plugin/httplink"
	_ "github.com/baetyl/baetyl/v2/plugin/pubsub"
)

func main() {
	cmd.Execute()
}
