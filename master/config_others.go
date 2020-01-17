// +build !linux

package master

import (
	"time"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/protocol/http"
)

// Config master init config
type Config struct {
	Mode   string          `yaml:"mode" json:"mode" default:"docker" validate:"regexp=^(native|docker)$"`
	Server http.ServerInfo `yaml:"server" json:"server" default:"{\"address\":\"tcp://127.0.0.1:50050\"}"`
	Logger logger.LogInfo  `yaml:"logger" json:"logger" default:"{\"path\":\"var/log/baetyl/baetyl.log\"}"`
	OTALog logger.LogInfo  `yaml:"otalog" json:"otalog" default:"{\"path\":\"var/db/baetyl/ota.log\",\"format\":\"json\"}"`
	Grace  time.Duration   `yaml:"grace" json:"grace" default:"30s"`
	SNFile string          `yaml:"snfile" json:"snfile"`
	Docker struct {
		APIVersion string `yaml:"api_version" json:"api_version" default:"1.38"`
	} `yaml:"docker" json:"docker"`

	// cache config file path
	File string
}
