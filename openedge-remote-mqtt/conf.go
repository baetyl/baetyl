package main

import (
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/trans/mqtt"
)

// Config config of module
type Config struct {
	module.Config `yaml:",inline" json:",inline"`
	Hub           mqtt.ClientConfig `yaml:"hub" json:"hub"`
	Remotes       []Remote          `yaml:"remotes" json:"remotes" default:"[]"`
	Rules         []Rule            `yaml:"rules" json:"rules" default:"[]"`
}

// Remote remote config
type Remote struct {
	Name              string `yaml:"name" json:"name" validate:"nonzero"`
	mqtt.ClientConfig `yaml:",inline" json:",inline"`
}

// Rule rule config
type Rule struct {
	ID  string `yaml:"id" json:"id"`
	Hub struct {
		Subscriptions []mqtt.Subscription `yaml:"subscriptions" json:"subscriptions" default:"[]"`
	} `yaml:"hub" json:"hub"`
	Remote struct {
		Name          string              `yaml:"name" json:"name" validate:"nonzero"`
		Subscriptions []mqtt.Subscription `yaml:"subscriptions" json:"subscriptions" default:"[]"`
	} `yaml:"remote" json:"remote"`
}
