package main

import (
	openedge "github.com/baidu/openedge/api/go"
)

// Config function module config
type Config struct {
	ImagePrefix string         `yaml:"prefix" json:"prefix"`
	Functions   []FunctionInfo `yaml:"functions" json:"functions" default:"[]"`
}

// FunctionInfo config
type FunctionInfo struct {
	Name      string             `yaml:"name" json:"name" validate:"nonzero"`
	Runtime   string             `yaml:"runtime" json:"runtime" validate:"nonzero"`
	Handler   string             `yaml:"handler" json:"handler" validate:"nonzero"`
	CodeDir   string             `yaml:"codedir" json:"codedir"`
	Env       map[string]string  `yaml:"env" json:"env"`
	Instance  Instance           `yaml:"instance" json:"instance"`
	Subscribe openedge.TopicInfo `yaml:"subscribe" json:"subscribe"`
	Publish   openedge.TopicInfo `yaml:"publish" json:"publish"`
}

// Instance instance config for function runtime module
type Instance struct {
	Min int `yaml:"min" json:"min" default:"0" validate:"min=0, max=100"`
	Max int `yaml:"max" json:"max" default:"1" validate:"min=1, max=100"`
	/* TODO
	IdleTime  time.Duration `yaml:"idletime" json:"idletime" default:"10m"`
	Timeout   time.Duration `yaml:"timeout" json:"timeout" default:"5m"`
	Resources Resources     `yaml:",inline"  json:",inline"`
	Message   struct {
		Length Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
	*/
}

// RuntimeInfo config
type RuntimeInfo struct {
	openedge.Config `yaml:",inline" json:",inline"`
	Subscribe       openedge.TopicInfo `yaml:"subscribe" json:"subscribe"`
	Publish         openedge.TopicInfo `yaml:"publish" json:"publish"`
	Name            string             `yaml:"name" json:"name" validate:"nonzero"`
	Handler         string             `yaml:"handler" json:"handler" validate:"nonzero"`
}
