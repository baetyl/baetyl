package memory

import (
	"github.com/baetyl/baetyl-go/v2/mq"
	"github.com/baetyl/baetyl-go/v2/mq/memory"
	v2plugin "github.com/baetyl/baetyl-go/v2/plugin"
	"github.com/baetyl/baetyl-go/v2/utils"

	"github.com/baetyl/baetyl/plugin"
)

func init() {
	v2plugin.RegisterFactory("defaultmq", NewMQ)
}

type defaultMQ struct {
	mq.MessageQueue
}

func NewMQ() (v2plugin.Plugin, error) {
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
