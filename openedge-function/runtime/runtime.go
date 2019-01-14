package runtime

import (
	"time"

	"github.com/baidu/openedge/utils"
)

// Config data
type Config struct {
	Server ServerInfo `yaml:"server" json:"server"`
	//Function function.FunctionInfo `yaml:"function" json:"function"`
}

// ClientInfo function runtime client config
type ClientInfo struct {
	ServerInfo `yaml:",inline" json:",inline"`
	Backoff    struct {
		Max time.Duration `yaml:"max" json:"max" default:"1m"`
	} `yaml:"backoff" json:"backoff"`
}

// ServerInfo function runtime server config
type ServerInfo struct {
	Address string        `yaml:"address" json:"address" validate:"nonzero"`
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	Message struct {
		Length Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
}

// Length length
type Length struct {
	Max int64 `yaml:"max" json:"max"`
}

// NewClientInfo create a new client config
func NewClientInfo(address string) ClientInfo {
	var cc ClientInfo
	utils.SetDefaults(&cc)
	cc.Address = address
	return cc
}
