package pubsub

import (
	goplugin "github.com/baetyl/baetyl-go/v2/plugin"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	"github.com/baetyl/baetyl-go/v2/utils"

	"github.com/baetyl/baetyl/v2/plugin"
)

func init() {
	goplugin.RegisterFactory("defaultpubsub", NewPubsub)
}

type defaultpb struct {
	pubsub.Pubsub
}

func NewPubsub() (goplugin.Plugin, error) {
	var cfg CloudConfig
	if err := utils.LoadYAML(plugin.ConfFile, &cfg); err != nil {
		return nil, err
	}

	res, err := pubsub.NewPubsub(cfg.Pubsub.Size)
	if err != nil {
		return nil, err
	}
	return &defaultpb{Pubsub: res}, nil
}
