package main

import (
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/trans/mqtt"
)

// Config config of module
type Config struct {
	module.Config `yaml:",inline" json:",inline"`
	Hub           mqtt.ClientConfig `yaml:"hub" json:"hub"`
	Remote        mqtt.ClientConfig `yaml:"remote" json:"remote"`
}
