package helper

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	v2plugin "github.com/baetyl/baetyl-go/v2/plugin"

	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/plugin"
)

const (
	TopicUpside   = "upside"
	TopicDownside = "downside"
)

//go:generate mockgen -destination=../mock/helper/helper.go -package=helper github.com/baetyl/baetyl/helper Helper

// Helper: used for message management between engine and sync
type Helper interface {
	plugin.MessageQueue
}

type helperImpl struct {
	plugin.MessageQueue
}

func NewHelper(cfg config.Config) (Helper, error) {
	mq, err := v2plugin.GetPlugin(cfg.Plugin.MQ)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &helperImpl{MessageQueue: mq.(plugin.MessageQueue)}, nil
}
