package config

import (
	"time"
)

// Function function config
type Function struct {
	Name    string `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9_-]{1\\,140}$"`
	Handler string `yaml:"handler" json:"handler" validate:"nonzero"`
	CodeDir string `yaml:"codedir" json:"codedir"`

	Instance Instance          `yaml:"instance" json:"instance"`
	Entry    string            `yaml:"entry" json:"entry"`
	Env      map[string]string `yaml:"env" json:"env"`
}

// Instance instance config for function runtime module
type Instance struct {
	Min       int           `yaml:"min" json:"min" default:"0" validate:"min=0, max=100"`
	Max       int           `yaml:"max" json:"max" default:"1" validate:"min=1, max=100"`
	IdleTime  time.Duration `yaml:"idletime" json:"idletime" default:"10m"`
	Timeout   time.Duration `yaml:"timeout" json:"timeout" default:"5m"`
	Resources Resources     `yaml:",inline"  json:",inline"`
	Message   struct {
		Length Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
}
