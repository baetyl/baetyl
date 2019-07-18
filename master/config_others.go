// +build !linux

package master

import (
	"time"

	"github.com/baidu/openedge/utils"
)

// Server server config
type Server struct {
	Address           string        `yaml:"address" json:"address" default:"tcp://127.0.0.1:50050"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout" default:"5m"`
	utils.Certificate `yaml:",inline" json:",inline"`
}
