package config

import (
	"time"
)

// RestartPolicies
const (
	RestartNo            = "no"
	RestartAlways        = "always"
	RestartOnFailure     = "on-failure"
	RestartUnlessStopped = "unless-stopped"
)

// Policy restart policy of a module
type Policy struct {
	Retry struct {
		Max int `yaml:"max" json:"max"`
	} `yaml:"retry" json:"retry"`
	Policy  string  `yaml:"policy" json:"policy" default:"always"`
	Backoff Backoff `yaml:"backoff" json:"backoff"`
}

// Backoff backoff
type Backoff struct {
	Min    time.Duration `yaml:"min" json:"min" default:"1s"`
	Max    time.Duration `yaml:"max" json:"max" default:"5m"`
	Factor float64       `yaml:"factor" json:"factor" default:"2"`
}
