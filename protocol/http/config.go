package http

import (
	"time"

	"github.com/baetyl/baetyl/utils"
)

// ClientInfo http client config
type ClientInfo struct {
	Address           string        `yaml:"address" json:"address"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout" default:"5m"`
	KeepAlive         time.Duration `yaml:"keepalive" json:"keepalive" default:"10m"`
	Username          string        `yaml:"username" json:"username"`
	Password          string        `yaml:"password" json:"password"`
	utils.Certificate `yaml:",inline" json:",inline"`
}
