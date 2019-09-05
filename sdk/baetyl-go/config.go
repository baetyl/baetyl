package baetyl

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

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
	ServiceNetworks map[string]ServiceNetwork `yaml:"networks" json:"networks"`
}

// ServiceNetwork specific network configuration of service
type ServiceNetwork struct {
	Aliases     []string `yaml:"aliases" json:"aliases"`
	Ipv4Address string   `yaml:"ipv4_address" json:"ipv4_address"`
}

// UnmarshalYAML customizes unmarshal
func (sn *NetworksInfo) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if sn.ServiceNetworks == nil {
		sn.ServiceNetworks = make(map[string]ServiceNetwork)
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
			sn.ServiceNetworks[name] = ServiceNetwork{}
		}
	case reflect.Map:
		return unmarshal(&sn.ServiceNetworks)
	default:
		return fmt.Errorf("failed to parse service network: unexpected type")
	}
	return nil
}

// ComposeAppConfig application configuration of compose
type ComposeAppConfig struct {
	// specifies the version of compose file
	Version string `yaml:"version" json:"version"`
	// specifies the app version of the application configuration
	AppVersion string `yaml:"app_version" json:"app_version"`
	// specifies the service information of the application
	Services map[string]ComposeService `yaml:"services" json:"services" default:"{}"`
	// specifies the storage volume information of the application
	Volumes map[string]ComposeVolume `yaml:"volumes" json:"volumes" default:"{}"`
	// specifies the network information of the applicaiton
	Networks map[string]ComposeNetwork `yaml:"networks" json:"networks" default:"{}"`
}

// ComposeService service configuration of compose
type ComposeService struct {
	// specifies the unique name of the service
	ContainerName string `yaml:"container_name" json:"container_name"`
	// specifies the hostname of the service
	Hostname string `yaml:"hostname" json:"hostname"`
	// specifies the image of the service, usually using the Docker image name
	Image string `yaml:"image" json:"image" validate:"nonzero"`
	// specifies the number of instances started
	Replica int `yaml:"replica" json:"replica" validate:"min=0"`
	// specifies the storage volumes that the service needs, map the storage volume to the directory in the container
	Volumes []ServiceVolume `yaml:"volumes" json:"volumes"`
	// specifies the network mode of the service
	NetworkMode string `yaml:"network_mode" json:"network_mode" validate:"regexp=^(bridge|host|none)?$"`
	// specifies the network that the service needs
	Networks NetworksInfo `yaml:"networks" json:"networks"`
	// specifies the port bindings which exposed by the service, only for Docker container mode
	Ports []string `yaml:"ports" json:"ports" default:"[]"`
	// specifies the device bindings which used by the service, only for Docker container mode
	Devices []string `yaml:"devices" json:"devices" default:"[]"`
	// specified other depended services
	DependsOn []string `yaml:"depends_on" json:"depends_on" default:"[]"`
	// specifies the startup arguments of the service program, but does not include `arg[0]`
	Command Command `yaml:"command" json:"command"`
	// specifies the environment variable of the service program
	Environment Environment `yaml:"environment" json:"environment" default:"{}"`
	// specifies the restart policy of the instance of the service
	Restart RestartPolicyInfo `yaml:"restart" json:"restart"`
	// specifies resource limits for a single instance of the service,  only for Docker container mode
	Resources Resources `yaml:"resources" json:"resources"`
	// specifies runtime to use, only for Docker container mode
	Runtime string `yaml:"runtime" json:"runtime"`
}

// ComposeVolume volume configuration of compose
type ComposeVolume struct {
	// specified driver for the storage volume
	Driver string `yaml:"driver" json:"driver" default:"local"`
	// specified driver options for the storage volume
	DriverOpts map[string]string `yaml:"driver_opts" json:"driver_opts" default:"{}"`
	// specified labels for the storage volume
	Labels map[string]string `yaml:"labels" json:"labels" default:"{}"`
}

// ComposeNetwork network configuration
type ComposeNetwork struct {
	// specifies driver for network
	Driver string `yaml:"driver" json:"driver" default:"bridge"`
	// specified driver options for network
	DriverOpts map[string]string `yaml:"driver_opts" json:"driver_opts" default:"{}"`
	// specifies labels to add metadata
	Labels map[string]string `yaml:"labels" json:"labels" default:"{}"`
}

// Environment environment
type Environment struct {
	Envs map[string]string `yaml:"envs" json:"envs" default:"{}"`
}

// UnmarshalYAML customize unmarshal
func (e *Environment) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if e.Envs == nil {
		e.Envs = make(map[string]string)
	}
	var envs interface{}
	err := unmarshal(&envs)
	if err != nil {
		return err
	}
	if envs == nil {
		return nil
	}
	switch reflect.ValueOf(envs).Kind() {
	case reflect.Slice:
		for _, env := range envs.([]interface{}) {
			envStr := env.(string)
			es := strings.Split(envStr, "=")
			if len(es) != 2 {
				return fmt.Errorf("environment format error")
			}
			e.Envs[es[0]] = es[1]
		}
	case reflect.Map:
		return unmarshal(&e.Envs)
	default:
		return fmt.Errorf("failed to parse environment: unexpected type")
	}
	return nil
}

// ServiceVolume specific volume configuration of service
type ServiceVolume struct {
	// specifies type of volume
	Type string `yaml:"type" json:"type" validate:"regexp=^(bind|volume)$"`
	// specifies source of volume
	Source string `yaml:"source" json:"source"`
	// specifies target of volume
	Target string `yaml:"target" json:"target"`
	// specifies if the volume is read-only
	ReadOnly bool `yaml:"read_only" json:"read_only"`
}

// UnmarshalYAML customize ServiceVolume unmarshal
func (sv *ServiceVolume) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var volume interface{}
	err := unmarshal(&volume)
	if err != nil {
		return err
	}
	if volume == nil {
		return nil
	}
	switch reflect.ValueOf(volume).Kind() {
	case reflect.String:
		volumeStr := volume.(string)
		info := strings.Split(volumeStr, ":")
		length := len(info)
		if length < 2 || length > 3 {
			return fmt.Errorf("servie volume format error")
		}
		sv.Source = info[0]
		sv.Target = info[1]
		if length == 3 && info[2] == "ro" {
			sv.ReadOnly = true
		}
		if match, _ := regexp.MatchString("^[a-zA-Z0-9][a-zA-Z0-9_.-]+$", sv.Source); match {
			sv.Type = "volume"
		} else {
			sv.Type = "bind"
		}
	case reflect.Map:
		type VolumeInfo ServiceVolume
		var volumeInfo VolumeInfo
		err := unmarshal(&volumeInfo)
		if err != nil {
			return err
		}
		*sv = ServiceVolume(volumeInfo)
	default:
		return fmt.Errorf("failed to parse service volume: unexpected type")
	}
	return nil
}

// Command command configuration of the service
type Command struct {
	Cmd []string `yaml:"cmd" json:"cmd" default:"[]"`
}

//UnmarshalYAML customize Command unmarshal
func (c *Command) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if c.Cmd == nil {
		c.Cmd = make([]string, 0)
	}
	var cmd interface{}
	err := unmarshal(&cmd)
	if err != nil {
		return err
	}
	switch reflect.ValueOf(cmd).Kind() {
	case reflect.String:
		c.Cmd = strings.Split(cmd.(string), " ")
	case reflect.Slice:
		return unmarshal(&c.Cmd)
	}
	return nil
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

// LoadComposeAppConfigCompatible load compose app config or old compatible config
func LoadComposeAppConfigCompatible(configFile string) (ComposeAppConfig, error) {
	var cfg ComposeAppConfig
	err := utils.LoadYAML(configFile, &cfg)
	if err != nil {
		var c AppConfig
		err = utils.LoadYAML(configFile, &c)
		if err != nil {
			return cfg, err
		}
		cfg = c.ToComposeAppConfig()
	}
	return cfg, nil
}
