package openedge

import (
	"time"

	"github.com/baidu/openedge/utils"
	units "github.com/docker/go-units"
)

// ServiceInfo of service
type ServiceInfo struct {
	Image     string            `yaml:"image" json:"image" validate:"nonzero"`
	Replica   int               `yaml:"replica" json:"replica" default:"1"`
	Expose    []string          `yaml:"expose" json:"expose" default:"[]"`
	Params    []string          `yaml:"params" json:"params" default:"[]"`
	Env       map[string]string `yaml:"env" json:"env" default:"{}"`
	Restart   RestartPolicyInfo `yaml:"restart" json:"restart"`
	Resources Resources         `yaml:"resources" json:"resources"`
	Mounts    []MountInfo       `yaml:"mounts" json:"mounts" default:"[]"`
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
	Min    time.Duration `yaml:"min" json:"min" default:"1s"`
	Max    time.Duration `yaml:"max" json:"max" default:"5m"`
	Factor float64       `yaml:"factor" json:"factor" default:"2"`
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

// MountInfo for mount
type MountInfo struct {
	Volume   string `yaml:"volume" json:"volume" validate:"nonzero"`
	Target   string `yaml:"target" json:"target" validate:"nonzero"`
	ReadOnly bool   `yaml:"readonly" json:"readonly" default:"false"`
}

// Config of service
type Config struct {
	Hub    MqttClientInfo `yaml:"hub" json:"hub"`
	Logger LogInfo        `yaml:"logger" json:"logger"`
}

// MqttClientInfo config
type MqttClientInfo struct {
	Address           string `yaml:"address" json:"address"`
	Username          string `yaml:"username" json:"username"`
	Password          string `yaml:"password" json:"password"`
	utils.Certificate `yaml:",inline" json:",inline"`
	ClientID          string        `yaml:"clientid" json:"clientid"`
	CleanSession      bool          `yaml:"cleansession" json:"cleansession" default:"false"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	Interval          time.Duration `yaml:"interval" json:"interval" default:"1m"`
	KeepAlive         time.Duration `yaml:"keepalive" json:"keepalive" default:"1m"`
	BufferSize        int           `yaml:"buffersize" json:"buffersize" default:"10"`
	ValidateSubs      bool          `yaml:"validatesubs" json:"validatesubs"`
	Subscriptions     []TopicInfo   `yaml:"subscriptions" json:"subscriptions" default:"[]"`
}

// HttpClientInfo http client config
type HttpClientInfo struct {
	Address           string        `yaml:"address" json:"address"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	KeepAlive         time.Duration `yaml:"keepalive" json:"keepalive" default:"30s"`
	Username          string        `yaml:"username" json:"username"`
	Password          string        `yaml:"password" json:"password"`
	utils.Certificate `yaml:",inline" json:",inline"`
}

// TopicInfo with topic and qos
type TopicInfo struct {
	Topic string `yaml:"topic" json:"topic" validate:"nonzero"`
	QoS   byte   `yaml:"qos" json:"qos" validate:"max=1"`
}

// LogInfo for logging
type LogInfo struct {
	Path    string `yaml:"path" json:"path" default:"var/log/openedge/service.log"`
	Level   string `yaml:"level" json:"level" default:"info" validate:"regexp=^(info|debug|warn|error)$"`
	Format  string `yaml:"format" json:"format" default:"text" validate:"regexp=^(text|json)$"`
	Console bool   `yaml:"console" json:"console" default:"false"`
	Age     struct {
		Max int `yaml:"max" json:"max" default:"15" validate:"min=1"`
	} `yaml:"age" json:"age"` // days
	Size struct {
		Max int `yaml:"max" json:"max" default:"50" validate:"min=1"`
	} `yaml:"size" json:"size"` // in MB
	Backup struct {
		Max int `yaml:"max" json:"max" default:"15" validate:"min=0"`
	} `yaml:"backup" json:"backup"`
}
