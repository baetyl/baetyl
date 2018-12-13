package runtime

import (
	"time"

	"github.com/baidu/openedge/config"
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/utils"
)

// Config runtime config
type Config struct {
	module.Config `yaml:",inline" json:",inline"`
	Server        ServerConfig   `yaml:"server" json:"server"`
	Function      FunctionConfig `yaml:"function" json:"function"`
}

// ClientConfig function runtime client config
type ClientConfig struct {
	ServerConfig `yaml:",inline" json:",inline"`
	Backoff      struct {
		Max time.Duration `yaml:"max" json:"max" default:"1m"`
	} `yaml:"backoff" json:"backoff"`
}

// ServerConfig function runtime server config
type ServerConfig struct {
	Address string        `yaml:"address" json:"address" validate:"nonzero"`
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	Message struct {
		Length config.Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
}

// FunctionConfig function config
type FunctionConfig struct {
	Name    string            `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9_-]{1\\,140}$"`
	Handler string            `yaml:"handler" json:"handler" validate:"nonzero"`
	CodeDir string            `yaml:"codedir" json:"codedir"`
	// Runtime string            `yaml:"runtime" json:"runtime"`
	Entry   string            `yaml:"entry" json:"entry"`
	// Env     map[string]string `yaml:"env" json:"env"`
}

// NewClientConfig create a new client config
func NewClientConfig(address string) ClientConfig {
	cc := ClientConfig{ServerConfig: ServerConfig{Address: address}}
	utils.SetDefaults(&cc)
	return cc
}
