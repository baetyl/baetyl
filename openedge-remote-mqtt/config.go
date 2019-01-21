package main

import (
	openedge "github.com/baidu/openedge/api/go"
)

// Config config of module
type Config struct {
	openedge.Config `yaml:",inline" json:",inline"`
	Remotes         []Remote `yaml:"remotes" json:"remotes" default:"[]"`
	Rules           []Rule   `yaml:"rules" json:"rules" default:"[]"`
}

// Remote remote config
type Remote struct {
	Name                    string `yaml:"name" json:"name" validate:"nonzero"`
	openedge.MqttClientInfo `yaml:",inline" json:",inline"`
}

// Rule rule config
type Rule struct {
	ID  string `yaml:"id" json:"id"`
	Hub struct {
		Subscriptions []openedge.TopicInfo `yaml:"subscriptions" json:"subscriptions" default:"[]"`
	} `yaml:"hub" json:"hub"`
	Remote struct {
		Name          string               `yaml:"name" json:"name" validate:"nonzero"`
		Subscriptions []openedge.TopicInfo `yaml:"subscriptions" json:"subscriptions" default:"[]"`
	} `yaml:"remote" json:"remote"`
}
