// +build !linux

package master

import (
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/protocol/http"
)

// Config master init config
type Config struct {
	Mode   string          `yaml:"mode" json:"mode" default:"docker" validate:"regexp=^(native|docker)$"`
	Server http.ServerInfo `yaml:"server" json:"server" default:"{\"address\":\"tcp://127.0.0.1:50050\"}"`
	Logger logger.LogInfo  `yaml:"logger" json:"logger" default:"{\"path\":\"var/log/openedge/openedge.log\"}"`
	OTALog logger.LogInfo  `yaml:"otalog" json:"otalog" default:"{\"path\":\"var/db/openedge/ota.log\",\"format\":\"json\"}"`
	Grace  time.Duration   `yaml:"grace" json:"grace" default:"30s"`

	// cache config file path
	File string
}
