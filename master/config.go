package master

import (
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/protocol/http"
)

// Config master init config
type Config struct {
	Mode   string          `yaml:"mode" json:"mode" default:"docker" validate:"regexp=^(native|docker)$"`
	Server http.ServerInfo `yaml:"server" json:"server"`
	Logger logger.LogInfo  `yaml:"logger" json:"logger"`
	Grace  time.Duration   `yaml:"grace" json:"grace" default:"30s"`
}
