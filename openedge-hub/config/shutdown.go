package config

import "time"

// Shutdown shutdown config
type Shutdown struct {
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"10m"`
}
