package http

import (
	"time"

	"github.com/baidu/openedge/trans"
)

// ClientConfig http client config
type ClientConfig struct {
	Address   string        `yaml:"address" json:"address"`
	Timeout   time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	KeepAlive time.Duration `yaml:"keepalive" json:"keepalive" default:"30s"`

	Username          string `yaml:"username" json:"username"`
	Password          string `yaml:"password" json:"password"`
	trans.Certificate `yaml:",inline" json:",inline"`
}

// ServerConfig http server config
type ServerConfig struct {
	Address           string        `yaml:"address" json:"address"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	trans.Certificate `yaml:",inline" json:",inline"`
}
