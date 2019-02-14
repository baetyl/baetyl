package engine

import (
	"time"

	units "github.com/docker/go-units"
)

// ServiceInfo of service
type ServiceInfo struct {
	Name      string            `yaml:"name" json:"name" validate:"nonzero"`
	Image     string            `yaml:"image" json:"image" validate:"nonzero"`
	Replica   int               `yaml:"replica" json:"replica" default:"1"`
	Expose    []string          `yaml:"expose" json:"expose" default:"[]"`
	Args      []string          `yaml:"args" json:"args" default:"[]"`
	Env       map[string]string `yaml:"env" json:"env" default:"{}"`
	Restart   RestartPolicyInfo `yaml:"restart" json:"restart"`
	Resources Resources         `yaml:"resources" json:"resources"`
	Datasets  []Dataset         `yaml:"datasets" json:"datasets" default:"[]"`
	Volumes   []MountInfo       `yaml:"volumes" json:"volumes" default:"[]"`
}

// Dataset for mount
type Dataset struct {
	Name     string `yaml:"name" json:"name" validate:"nonzero"`
	Version  string `yaml:"version" json:"version"`
	Target   string `yaml:"target" json:"target" validate:"nonzero"`
	ReadOnly bool   `yaml:"readonly" json:"readonly" default:"false"`
}

// MountInfo for mount
type MountInfo struct {
	Volume   string `yaml:"volume" json:"volume" validate:"nonzero"`
	Target   string `yaml:"target" json:"target" validate:"nonzero"`
	ReadOnly bool   `yaml:"readonly" json:"readonly" default:"false"`
}

// RestartPolicies
const (
	RestartNo        = "no"
	RestartAlways    = "always"
	RestartOnFailure = "on-failure"
)

// RestartPolicyInfo holds the policy of a module
type RestartPolicyInfo struct {
	Retry struct {
		Max int `yaml:"max" json:"max"`
	} `yaml:"retry" json:"retry"`
	Policy  string      `yaml:"policy" json:"policy" default:"always"`
	Backoff BackoffInfo `yaml:"backoff" json:"backoff"`
}

// BackoffInfo holds backoff value
type BackoffInfo struct {
	Min    time.Duration `yaml:"min" json:"min" default:"1s" validate:"min=1000000000"`
	Max    time.Duration `yaml:"max" json:"max" default:"5m" validate:"min=1000000000"`
	Factor float64       `yaml:"factor" json:"factor" default:"2" validate:"min=1"`
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

type memory struct {
	Limit string `yaml:"limit" json:"limit"`
	Swap  string `yaml:"swap" json:"swap"`
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
