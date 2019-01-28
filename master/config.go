package master

import (
	"time"

	openedge "github.com/baidu/openedge/api/go"
)

// Config master config
type Config struct {
	Mode     string                 `yaml:"mode" json:"mode" default:"docker" validate:"regexp=^(native|docker)$"`
	Server   string                 `yaml:"server" json:"server"`
	Services []openedge.ServiceInfo `yaml:"services" json:"services" default:"[]"`
	Grace    time.Duration          `yaml:"grace" json:"grace" default:"30s"`
	Logger   openedge.LogInfo       `yaml:"logger" json:"logger"`
}

// DynamicConfig for reload
type DynamicConfig struct {
	Version  string                 `yaml:"version" json:"version"`
	Services []openedge.ServiceInfo `yaml:"services" json:"services" default:"[]"`
	Grace    time.Duration          `yaml:"grace" json:"grace" default:"30s"`
	Logger   openedge.LogInfo       `yaml:"logger" json:"logger"`
}
