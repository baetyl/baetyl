package master

import (
	"time"

	"github.com/baidu/openedge/logger"
)

// Config master init config
type Config struct {
	Mode   string         `yaml:"mode" json:"mode" default:"docker" validate:"regexp=^(native|docker)$"`
	Server Server         `yaml:"server" json:"server"`
	Logger logger.LogInfo `yaml:"logger" json:"logger"`
	Grace  time.Duration  `yaml:"grace" json:"grace" default:"30s"`
}
