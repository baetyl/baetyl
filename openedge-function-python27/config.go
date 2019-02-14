package main

import (
	"github.com/baidu/openedge/protocol/mqtt"
	"github.com/baidu/openedge/sdk-go/openedge"
)

// Config of runtime
type Config struct {
	openedge.Config `yaml:",inline" json:",inline"`
	Subscribe       mqtt.TopicInfo `yaml:"subscribe" json:"subscribe"`
	Publish         mqtt.TopicInfo `yaml:"publish" json:"publish"`
	Name            string         `yaml:"name" json:"name" validate:"nonzero"`
	Handler         string         `yaml:"handler" json:"handler" validate:"nonzero"`
}
