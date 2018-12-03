package config

import "time"

// Status status config
type Status struct {
	Logging struct {
		Enable   bool          `yaml:"enable" json:"enable"`
		Interval time.Duration `yaml:"interval" json:"interval" default:"1m"`
	} `yaml:"logging" json:"logging"`
}
