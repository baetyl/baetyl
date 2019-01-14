package http

import (
	"time"

	"github.com/baidu/openedge/utils"
)

// ServerInfo of http server
type ServerInfo struct {
	Address           string        `yaml:"address" json:"address"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	utils.Certificate `yaml:",inline" json:",inline"`
}
