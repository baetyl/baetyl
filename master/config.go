package master

import (
	"time"

	"github.com/baidu/openedge/agent"
	"github.com/baidu/openedge/module/config"
)

// Config master config
type Config struct {
	Mode    string            `yaml:"mode" json:"mode" default:"docker" validate:"regexp=^(native|docker)$"`
	Version string            `yaml:"version" json:"version"`
	Grace   time.Duration     `yaml:"grace" json:"grace" default:"30s"`
	API     config.HTTPServer `yaml:"api" json:"api"`
	Cloud   agent.Config      `yaml:"cloud" json:"cloud"`
	Modules []config.Module   `yaml:"modules" json:"modules" default:"[]"`
	Logger  config.Logger     `yaml:"logger" json:"logger"`
}
