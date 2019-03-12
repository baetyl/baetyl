package openedge

import (
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/protocol/mqtt"
	"github.com/baidu/openedge/utils"
	units "github.com/docker/go-units"
)

// ServiceConfig base config of service
type ServiceConfig struct {
	Name   string          `yaml:"name" json:"name"`
	Hub    mqtt.ClientInfo `yaml:"hub" json:"hub"`
	Logger logger.LogInfo  `yaml:"logger" json:"logger"`
}

// AppConfig dynamic config of application
type AppConfig struct {
	Version  string        `yaml:"version" json:"version"`
	Services []ServiceInfo `yaml:"services" json:"services" default:"[]"`
	Volumes  []VolumeInfo  `yaml:"volumes" json:"volumes" default:"[]"`
}

// ServiceInfo of service
type ServiceInfo struct {
	Name      string            `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9][a-zA-Z0-9_-]{0\\,31}$"`
	Image     string            `yaml:"image" json:"image" validate:"nonzero"`
	Replica   int               `yaml:"replica" json:"replica"`
	Mounts    []MountInfo       `yaml:"mounts" json:"mounts" default:"[]"`
	Ports     []string          `yaml:"ports" json:"ports" default:"[]"`
	Args      []string          `yaml:"args" json:"args" default:"[]"`
	Env       map[string]string `yaml:"env" json:"env" default:"{}"`
	Restart   RestartPolicyInfo `yaml:"restart" json:"restart"`
	Resources Resources         `yaml:"resources" json:"resources"`
}

// VolumeInfo volume info
type VolumeInfo struct {
	Name     string `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9][a-zA-Z0-9_-]{0\\,31}$"`
	Path     string `yaml:"path" json:"path" validate:"nonzero"`
	ReadOnly bool   `yaml:"readonly" json:"readonly"`
	Meta     struct {
		URL string `yaml:"url" json:"url"`
		MD5 string `yaml:"md5" json:"md5"`
	} `yaml:"meta" json:"meta"`
}

// MountInfo volume mount info
type MountInfo struct {
	Name     string `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9-_]{1\\,32}$"`
	Path     string `yaml:"path" json:"path" validate:"nonzero"`
	ReadOnly bool   `yaml:"readonly" json:"readonly"`
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

/* functions */

// FunctionClientConfig functions client config
type FunctionClientConfig struct {
	Address string `yaml:"address" json:"address" validate:"nonzero"`
	Message struct {
		Length utils.Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
	Backoff struct {
		Max time.Duration `yaml:"max" json:"max" default:"1m"`
	} `yaml:"backoff" json:"backoff"`
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
}

// FunctionServerConfig functions server config
type FunctionServerConfig struct {
	Address string `yaml:"address" json:"address" validate:"nonzero"`
	Message struct {
		Length utils.Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
}

// Thread  struct {
// 	Workers struct {
// 		Max int `yaml:"max" json:"max"`
// 	} `yaml:"workers" json:"workers"`
// } `yaml:"thread" json:"thread"`
// Concurrent struct {
// 	RPCs struct {
// 		Max int `yaml:"max" json:"max"`
// 	} `yaml:"rpcs" json:"rpcs"`
// } `yaml:"concurrent" json:"concurrent"`
