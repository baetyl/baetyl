package main

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/module/config"
)

// Config function module config
type Config struct {
	config.Module `yaml:",inline" json:",inline"`
	API           config.HTTPClient `yaml:"api" json:"api"`
	Hub           config.MQTTClient `yaml:"hub" json:"hub"`
	Rules         []Rule            `yaml:"rules" json:"rules" default:"[]"`
	Functions     []config.Function `yaml:"functions" json:"functions" default:"[]"`
}

// Rule function rule config
type Rule struct {
	ID        string `yaml:"id" json:"id"`
	Subscribe struct {
		Topic string     `yaml:"topic" json:"topic" validate:"nonzero"`
		QOS   packet.QOS `yaml:"qos" json:"qos" default:"0" validate:"min=0, max=1"`
	} `yaml:"subscribe" json:"subscribe"`
	Compute struct {
		Function string `yaml:"function" json:"function" validate:"nonzero"`
	} `yaml:"compute" json:"compute"`
	Publish struct {
		Topic string     `yaml:"topic" json:"topic" validate:"nonzero"`
		QOS   packet.QOS `yaml:"qos" json:"qos" default:"0" validate:"min=0, max=1"`
	} `yaml:"publish" json:"publish"`
}
