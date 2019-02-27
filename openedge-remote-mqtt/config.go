package main

import (
	"github.com/baidu/openedge/protocol/mqtt"
)

// Config config of module
type Config struct {
	Remotes []Remote `yaml:"remotes" json:"remotes" default:"[]"`
	Rules   []Rule   `yaml:"rules" json:"rules" default:"[]"`
}

// Remote remote config
type Remote struct {
	Name            string `yaml:"name" json:"name" validate:"nonzero"`
	mqtt.ClientInfo `yaml:",inline" json:",inline"`
}

// Rule rule config
type Rule struct {
	Hub struct {
		ClientID      string           `yaml:"clientid" json:"clientid"`
		Subscriptions []mqtt.TopicInfo `yaml:"subscriptions" json:"subscriptions" default:"[]"`
	} `yaml:"hub" json:"hub"`
	Remote struct {
		Name          string           `yaml:"name" json:"name" validate:"nonzero"`
		ClientID      string           `yaml:"clientid" json:"clientid"`
		Subscriptions []mqtt.TopicInfo `yaml:"subscriptions" json:"subscriptions" default:"[]"`
	} `yaml:"remote" json:"remote"`
}
