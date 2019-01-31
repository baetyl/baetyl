package http

import (
	"time"

	"github.com/baidu/openedge/utils"
)

// ServerInfo http server config
type ServerInfo struct {
	Address           string        `yaml:"address" json:"address"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	KeepAlive         time.Duration `yaml:"keepalive" json:"keepalive" default:"60s"`
	utils.Certificate `yaml:",inline" json:",inline"`
}

// ClientInfo http client config
type ClientInfo struct {
	Address           string        `yaml:"address" json:"address"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	KeepAlive         time.Duration `yaml:"keepalive" json:"keepalive" default:"60s"`
	Username          string        `yaml:"username" json:"username"`
	Password          string        `yaml:"password" json:"password"`
	utils.Certificate `yaml:",inline" json:",inline"`
}
