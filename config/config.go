package config

import (
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/trans/http"
	"github.com/baidu/openedge/trans/mqtt"
	units "github.com/docker/go-units"
)

// Module define the meta data of a module
type Module struct {
	module.Config `yaml:",inline" json:",inline"`
	Entry         string            `yaml:"entry" json:"entry"`
	Restart       module.Policy     `yaml:"restart" json:"restart"`
	Expose        []string          `yaml:"expose" json:"expose" default:"[]"`
	Params        []string          `yaml:"params" json:"params" default:"[]"`
	Env           map[string]string `yaml:"env" json:"env" default:"{}"`
	Resources     module.Resources  `yaml:"resources" json:"resources"`
}

// Cloud cloud config
type Cloud struct {
	mqtt.ClientConfig `yaml:",inline" json:",inline"`
	OpenAPI           http.ClientConfig `yaml:"openapi" json:"openapi"`
}

// Master master config
type Master struct {
	Mode    string            `yaml:"mode" json:"mode" default:"docker" validate:"regexp=^(native|docker)$"`
	Version string            `yaml:"version" json:"version"`
	Grace   time.Duration     `yaml:"grace" json:"grace" default:"30s"`
	API     http.ServerConfig `yaml:"api" json:"api"`
	Cloud   Cloud             `yaml:"cloud" json:"cloud"`
	Modules []Module          `yaml:"modules" json:"modules" default:"[]"`
	Logger  logger.Config     `yaml:"logger" json:"logger"`
}

// Length length
type Length struct {
	Max int64 `yaml:"max" json:"max"`
}

// UnmarshalYAML customizes unmarshal
func (l *Length) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ls length
	err := unmarshal(&ls)
	if err != nil {
		return err
	}
	if ls.Max != "" {
		l.Max, err = units.RAMInBytes(ls.Max)
		if err != nil {
			return err
		}
	}
	return nil
}

type length struct {
	Max string `yaml:"max" json:"max"`
}
