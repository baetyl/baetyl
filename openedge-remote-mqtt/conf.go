package main

import (
	"github.com/baidu/openedge/module/config"
)

// Config config of module
type Config struct {
	config.Module `yaml:",inline" json:",inline"`
	Hub           config.MQTTClient `yaml:"hub" json:"hub"`
	Remotes       []Remote          `yaml:"remotes" json:"remotes" default:"[]"`
	Rules         []Rule            `yaml:"rules" json:"rules" default:"[]"`
}

// Remote remote config
type Remote struct {
	Name              string `yaml:"name" json:"name" validate:"nonzero"`
	config.MQTTClient `yaml:",inline" json:",inline"`
}

// Rule rule config
type Rule struct {
	ID  string `yaml:"id" json:"id"`
	Hub struct {
		Subscriptions []config.Subscription `yaml:"subscriptions" json:"subscriptions" default:"[]"`
	} `yaml:"hub" json:"hub"`
	Remote struct {
		Name          string                `yaml:"name" json:"name" validate:"nonzero"`
		Subscriptions []config.Subscription `yaml:"subscriptions" json:"subscriptions" default:"[]"`
	} `yaml:"remote" json:"remote"`
}
