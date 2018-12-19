package config

import (
	"time"

	"github.com/baidu/openedge/module/utils"
)

// HTTPClient http client config
type HTTPClient struct {
	Address   string        `yaml:"address" json:"address"`
	Timeout   time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	KeepAlive time.Duration `yaml:"keepalive" json:"keepalive" default:"30s"`

	Username          string `yaml:"username" json:"username"`
	Password          string `yaml:"password" json:"password"`
	utils.Certificate `yaml:",inline" json:",inline"`
}

// HTTPServer http server config
type HTTPServer struct {
	Address           string        `yaml:"address" json:"address"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	utils.Certificate `yaml:",inline" json:",inline"`
}
