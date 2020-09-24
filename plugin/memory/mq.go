package memory

import (
	"github.com/baetyl/baetyl-go/v2/mq"
	"github.com/baetyl/baetyl-go/v2/mq/memory"
	goplugin "github.com/baetyl/baetyl-go/v2/plugin"
	"github.com/baetyl/baetyl-go/v2/utils"

	"github.com/baetyl/baetyl/v2/plugin"
)

func init() {
	goplugin.RegisterFactory("defaultmq", NewMQ)
}

type defaultMQ struct {
	mq.MessageQueue
}

func NewMQ() (goplugin.Plugin, error) {
	var cfg CloudConfig
	if err := utils.LoadYAML(plugin.ConfFile, &cfg); err != nil {
		return nil, err
	}

	res, err := memory.NewMQ(cfg.MQ.Size, cfg.MQ.Duration)
	if err != nil {
		return nil, err
	}
	return &defaultMQ{MessageQueue: res}, nil
}
