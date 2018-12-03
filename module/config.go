package module

import (
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/docker/go-units"
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

// Config module config
type Config struct {
	// ID     string        `yaml:"id" json:"id"`
	Name   string        `yaml:"name" json:"name" validate:"nonzero"`
	Mark   string        `yaml:"mark" json:"mark"`
	Logger logger.Config `yaml:"logger" json:"logger"`
}

// Resources resources config
type Resources struct {
	CPU    CPU    `yaml:"cpu" json:"cpu"`
	Pids   Pids   `yaml:"pids" json:"pids"`
	Memory Memory `yaml:"memory" json:"memory"`
}

// CPU cpu config
type CPU struct {
	Cpus    float64 `yaml:"cpus" json:"cpus"`
	SetCPUs string  `yaml:"setcpus" json:"setcpus"`
}

// Pids pids config
type Pids struct {
	Limit int64 `yaml:"limit" json:"limit"`
}

// Memory memory config
type Memory struct {
	Limit int64 `yaml:"limit" json:"limit"`
	Swap  int64 `yaml:"swap" json:"swap"`
}

// UnmarshalYAML customizes unmarshal
func (m *Memory) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ms memory
	err := unmarshal(&ms)
	if err != nil {
		return err
	}
	if ms.Limit != "" {
		m.Limit, err = units.RAMInBytes(ms.Limit)
		if err != nil {
			return err
		}
	}
	if ms.Swap != "" {
		m.Swap, err = units.RAMInBytes(ms.Swap)
		if err != nil {
			return err
		}
	}
	return nil
}

type memory struct {
	Limit string `yaml:"limit" json:"limit"`
	Swap  string `yaml:"swap" json:"swap"`
}
