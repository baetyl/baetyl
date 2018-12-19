package config

import (
	"time"

	"github.com/baidu/openedge/module/utils"
)

// Runtime runtime config
type Runtime struct {
	Module   `yaml:",inline" json:",inline"`
	Server   RuntimeServer `yaml:"server" json:"server"`
	Function Function      `yaml:"function" json:"function"`
}

// RuntimeClient function runtime client config
type RuntimeClient struct {
	RuntimeServer `yaml:",inline" json:",inline"`
	Backoff       struct {
		Max time.Duration `yaml:"max" json:"max" default:"1m"`
	} `yaml:"backoff" json:"backoff"`
}

// RuntimeServer function runtime server config
type RuntimeServer struct {
	Address string        `yaml:"address" json:"address" validate:"nonzero"`
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	Message struct {
		Length Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
}

// NewRuntimeClient create a new client config
func NewRuntimeClient(address string) RuntimeClient {
	var cc RuntimeClient
	utils.SetDefaults(&cc)
	cc.Address = address
	return cc
}
