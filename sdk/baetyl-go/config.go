package baetyl

import (
	"time"
	"reflect"
	"fmt"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/protocol/mqtt"
	"github.com/baetyl/baetyl/utils"
	units "github.com/docker/go-units"
)

// ServiceConfig base config of service
type ServiceConfig struct {
	Hub    mqtt.ClientInfo `yaml:"hub" json:"hub"`
	Logger logger.LogInfo  `yaml:"logger" json:"logger"`
}

// AppConfig application configuration
type AppConfig struct {
	// specifies the version of the application configuration
	Version string `yaml:"version" json:"version"`
	// specifies the service information of the application
	Services []ServiceInfo `yaml:"services" json:"services" default:"[]"`
	// specifies the storage volume information of the application
	Volumes []VolumeInfo `yaml:"volumes" json:"volumes" default:"[]"`
	// specifies the network information of the application
	Networks map[string]NetworkInfo `yaml:"networks" json:"networks" default:"{}"`
}

// ServiceInfo service configuration
type ServiceInfo struct {
	// specifies the unique name of the service
	Name string `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9][a-zA-Z0-9_-]{0\\,63}$"`
	// specifies the image of the service, usually using the Docker image name
	Image string `yaml:"image" json:"image" validate:"nonzero"`
	// specifies the number of instances started
	Replica int `yaml:"replica" json:"replica" validate:"min=0"`
	// specifies the storage volumes that the service needs, map the storage volume to the directory in the container
	Mounts []MountInfo `yaml:"mounts" json:"mounts" default:"[]"`
	// specifies the network that the service used
	Networks NetworksInfo `yaml:"networks" json:"networks"`
	// specifies the network mode of the service
	NetworkMode string `yaml:"network_mode" json:"network_mode" validate:"regexp=^(bridge|host|none)?$"`
	// specifies the port bindings which exposed by the service, only for Docker container mode
	Ports []string `yaml:"ports" json:"ports" default:"[]"`
	// specifies the device bindings which used by the service, only for Docker container mode
	Devices []string `yaml:"devices" json:"devices" default:"[]"`
	// specifies the startup arguments of the service program, but does not include `arg[0]`
	Args []string `yaml:"args" json:"args" default:"[]"`
	// specifies the environment variable of the service program
	Env map[string]string `yaml:"env" json:"env" default:"{}"`
	// specifies the restart policy of the instance of the service
	Restart RestartPolicyInfo `yaml:"restart" json:"restart"`
	// specifies resource limits for a single instance of the service,  only for Docker container mode
	Resources Resources `yaml:"resources" json:"resources"`
	// specifies runtime to use, only for Docker container mode
	Runtime string `yaml:"runtime" json:"runtime"`
}

// VolumeInfo storage volume configuration
type VolumeInfo struct {
	// specifies a unique name for the storage volume
	Name string `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9][a-zA-Z0-9_-]{0\\,63}$"`
	// specifies the directory where the storage volume is on the host
	Path string `yaml:"path" json:"path" validate:"nonzero"`
	// specifies the metadata of the storage volume
	Meta struct {
		URL     string `yaml:"url" json:"url"`
		MD5     string `yaml:"md5" json:"md5"`
		Version string `yaml:"version" json:"version"`
	} `yaml:"meta" json:"meta"`
}

// NetworkInfo network configuration
type NetworkInfo struct {
	// specifies driver for network
	Driver string `yaml:"driver" json:"driver" default:"bridge"`
	// specified driver options for network
	DriverOpts map[string]string `yaml:"driver_opts" json:"driver_opts"`
	// specifies labels to add metadata
	Labels map[string]string `yaml:"labels" json:"labels"`
}

// MountInfo storage volume mapping configuration
type MountInfo struct {
	// specifies the name of the mapped storage volume
	Name string `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9][a-zA-Z0-9_-]{0\\,63}$"`
	// specifies the directory where the storage volume is in the container
	Path string `yaml:"path" json:"path" validate:"nonzero"`
	// specifies the operation permission of the storage volume, read-only or writable
	ReadOnly bool `yaml:"readonly" json:"readonly"`
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

/* function */

// FunctionClientConfig function client config
type FunctionClientConfig struct {
	Address string `yaml:"address" json:"address"`
	Message struct {
		Length utils.Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
	Backoff struct {
		Max time.Duration `yaml:"max" json:"max" default:"1m"`
	} `yaml:"backoff" json:"backoff"`
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
}

// FunctionServerConfig function server config
type FunctionServerConfig struct {
	Address string        `yaml:"address" json:"address"`
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"2m"`
	Message struct {
		Length utils.Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
	Concurrent struct {
		Max uint32 `yaml:"max" json:"max"`
	} `yaml:"concurrent" json:"concurrent"`
	// for python function server
	Workers struct {
		Max uint32 `yaml:"max" json:"max"`
	} `yaml:"workers" json:"workers"`
	utils.Certificate `yaml:",inline" json:",inline"`
}


 // NetworksInfo network configurations of service
type NetworksInfo struct {
	ServiceNetworkInfos map[string]ServiceNetworkInfo `yaml:"networks" json:"networks"`
}

// ServiceNetworkInfo specific network configuration of service
type ServiceNetworkInfo struct {
	Aliases []string `yaml:"aliases" json:"aliases"`
	Ipv4Address string `yaml:"ipv4_address" json:"ipv4_address"`
}

// UnmarshalYAML customizes unmarshal
func (sn *NetworksInfo) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if sn.ServiceNetworkInfos == nil {
		sn.ServiceNetworkInfos = make(map[string]ServiceNetworkInfo)
	}
	var networks interface{}
	err := unmarshal(&networks)
	if err != nil {
		return err
	}
	switch reflect.ValueOf(networks).Kind() {
		case reflect.Slice:
			for _, item := range networks.([]interface{}) {
				name := item.(string)
				sn.ServiceNetworkInfos[name] = ServiceNetworkInfo{}
			}
		case reflect.Map:
			err = unmarshal(&sn.ServiceNetworkInfos)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("parse service network error")
	}
	return nil
}