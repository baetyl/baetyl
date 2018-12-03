package function

import (
	"time"

	"github.com/baidu/openedge/config"
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/trans/http"
	"github.com/baidu/openedge/trans/mqtt"
)

// Config function module config
type Config struct {
	module.Config `yaml:",inline" json:",inline"`
	API           http.ClientConfig `yaml:"api" json:"api"`
	Hub           mqtt.ClientConfig `yaml:"hub" json:"hub"`
	Rules         []Rule            `yaml:"rules" json:"rules" default:"[]"`
	Functions     []FunctionConfig  `yaml:"functions" json:"functions" default:"[]"`
}

// Rule function rule config
type Rule struct {
	ID        string `yaml:"id" json:"id"`
	Subscribe struct {
		Topic string `yaml:"topic" json:"topic" validate:"nonzero"`
		QOS   byte   `yaml:"qos" json:"qos" default:"0" validate:"min=0, max=1"`
	} `yaml:"subscribe" json:"subscribe"`
	Compute struct {
		Function string `yaml:"function" json:"function" validate:"nonzero"`
	} `yaml:"compute" json:"compute"`
	Publish struct {
		Topic string `yaml:"topic" json:"topic" validate:"nonzero"`
		QOS   byte   `yaml:"qos" json:"qos" default:"0" validate:"min=0, max=1"`
	} `yaml:"publish" json:"publish"`
}

// FunctionConfig function config
type FunctionConfig struct {
	Name     string            `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9_-]{1\\,140}$"`
	Entry    string            `yaml:"entry" json:"entry" validate:"nonzero"`
	Handler  string            `yaml:"handler" json:"handler" validate:"nonzero"`
	CodeDir  string            `yaml:"codedir" json:"codedir"`
	Env      map[string]string `yaml:"env" json:"env"`
	Instance Instance          `yaml:"instance" json:"instance"`
}

// Instance instance config
type Instance struct {
	Min       int              `yaml:"min" json:"min" default:"0" validate:"min=0, max=100"`
	Max       int              `yaml:"max" json:"max" default:"1" validate:"min=1, max=100"`
	IdleTime  time.Duration    `yaml:"idletime" json:"idletime" default:"10m"`
	Timeout   time.Duration    `yaml:"timeout" json:"timeout" default:"5m"`
	Resources module.Resources `yaml:",inline"  json:",inline"`
	Message   struct {
		Length config.Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
}
