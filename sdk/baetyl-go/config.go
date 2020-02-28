package baetyl

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	// specifies name of the application
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// specifies the app version of the application configuration
	AppVersion string `yaml:"app_version,omitempty" json:"app_version,omitempty"`
	// specifies the service information of the application
	Services map[string]ComposeService `yaml:"services,omitempty" json:"services,omitempty" default:"{}"`
	// specifies the storage volume information of the application
	Volumes map[string]ComposeVolume `yaml:"volumes,omitempty" json:"volumes,omitempty" default:"{}"`
	// specifies the network information of the application
	Networks map[string]ComposeNetwork `yaml:"networks,omitempty" json:"networks,omitempty" default:"{}"`
}

// ComposeService service configuration of compose
type ComposeService struct {
	// specifies the unique name of the service
	ContainerName string `yaml:"container_name,omitempty" json:"container_name,omitempty"`
	// specifies the hostname of the service
	Hostname string `yaml:"hostname,omitempty" json:"hostname,omitempty"`
	// specifies the image of the service, usually using the Docker image name
	Image string `yaml:"image,omitempty" json:"image,omitempty" validate:"nonzero"`
	// specifies the number of instances started
	Replica int `yaml:"replica,omitempty" json:"replica,omitempty" validate:"min=0"`
	// specifies the storage volumes that the service needs, map the storage volume to the directory in the container
	Volumes []ServiceVolume `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	// specifies the network mode of the service
	NetworkMode string `yaml:"network_mode,omitempty" json:"network_mode,omitempty" validate:"regexp=^(bridge|host|none)?$"`
	// specifies the network that the service needs
	Networks NetworksInfo `yaml:"networks,omitempty" json:"networks,omitempty"`
	// specifies the port bindings which exposed by the service, only for Docker container mode
	Ports []string `yaml:"ports,omitempty" json:"ports,omitempty" default:"[]"`
	// specifies the device bindings which used by the service, only for Docker container mode
	Devices []string `yaml:"devices,omitempty" json:"devices,omitempty" default:"[]"`
	// specified other depended services
	DependsOn []string `yaml:"depends_on,omitempty" json:"depends_on,omitempty" default:"[]"`
	// specifies the startup arguments of the service program, but does not include `arg[0]`
	Command Command `yaml:"command,omitempty" json:"command,omitempty" default:"{}"`
	// specifies the entrypoint of the service program
	Entrypoint Entrypoint `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty" default:"{}"`
	// specifies the environment variable of the service program
	Environment Environment `yaml:"environment,omitempty" json:"environment,omitempty" default:"{}"`
	// specifies the restart policy of the instance of the service
	Restart RestartPolicyInfo `yaml:"restart,omitempty" json:"restart,omitempty"`
	// specifies resource limits for a single instance of the service,  only for Docker container mode
	Resources Resources `yaml:"resources,omitempty" json:"resources,omitempty"`
	// specifies runtime to use, only for Docker container mode
	Runtime string `yaml:"runtime,omitempty" json:"runtime,omitempty"`
}

// ComposeVolume volume configuration of compose
type ComposeVolume struct {
	// specified driver for the storage volume
	Driver string `yaml:"driver,omitempty" json:"driver,omitempty" default:"local"`
	// specified driver options for the storage volume
	DriverOpts map[string]string `yaml:"driver_opts,omitempty" json:"driver_opts,omitempty" default:"{}"`
	// specified labels for the storage volume
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty" default:"{}"`
}

// ComposeNetwork network configuration
type ComposeNetwork struct {
	// specifies driver for network
	Driver string `yaml:"driver,omitempty" json:"driver,omitempty" default:"bridge"`
	// specified driver options for network
	DriverOpts map[string]string `yaml:"driver_opts,omitempty" json:"driver_opts,omitempty" default:"{}"`
	// specifies labels to add metadata
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty" default:"{}"`
}

// Environment environment
type Environment struct {
	Envs map[string]string `yaml:"envs" json:"envs" default:"{}"`
}

func (e Environment) MarshalYAML() (interface{}, error) {
	return e.Envs, nil
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

func (e *Environment) UnmarshalJSON(b []byte) error {
	if e.Envs == nil {
		e.Envs = make(map[string]string)
	}
	var envs interface{}
	err := json.Unmarshal(b, &envs)
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
		return json.Unmarshal(b, &e.Envs)
	default:
		return fmt.Errorf("failed to parse environment: unexpected type")
	}
	return nil
}

func (e Environment) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Envs)
}

// ServiceVolume specific volume configuration of service
type ServiceVolume struct {
	// specifies type of volume
	Type string `yaml:"type,omitempty" json:"type,omitempty" validate:"regexp=^(volume|bind)?$"`
	// specifies source of volume
	Source string `yaml:"source,omitempty" json:"source,omitempty"`
	// specifies target of volume
	Target string `yaml:"target,omitempty" json:"target,omitempty"`
	// specifies if the volume is read-only
	ReadOnly bool `yaml:"read_only,omitempty" json:"read_only,omitempty"`
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

// MarshalYAML customize ServiceVolume marshal
func (sv ServiceVolume) MarshalYAML() (interface{}, error) {
	res := sv.Source + ":" + sv.Target
	if sv.ReadOnly {
		res += ":ro"
	}
	return res, nil
}

// Command command configuration of the service
type Command struct {
	Cmd []string `yaml:"cmd" json:"cmd" default:"[]"`
}

func (c Command) MarshalYAML() (interface{}, error) {
	return c.Cmd, nil
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

//UnmarshalJSON customize Command unmarshal
func (c *Command) UnmarshalJSON(b []byte) error {
	if c.Cmd == nil {
		c.Cmd = make([]string, 0)
	}
	var cmd interface{}
	err := json.Unmarshal(b, &cmd)
	if err != nil {
		return err
	}
	switch reflect.ValueOf(cmd).Kind() {
	case reflect.String:
		c.Cmd = strings.Split(cmd.(string), " ")
	case reflect.Slice:
		return json.Unmarshal(b, &c.Cmd)
	}
	return nil
}

func (c Command) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Cmd)
}

// Entrypoint entrypoint configuration of the service
type Entrypoint struct {
	Entry []string `yaml:"entry" json:"entry" default:"[]"`
}

// MarshalYAML customize Entrypoint marshal
func (e Entrypoint) MarshalYAML() (interface{}, error) {
	return e.Entry, nil
}

//UnmarshalYAML customize Entrypoint unmarshal
func (e *Entrypoint) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if e.Entry == nil {
		e.Entry = make([]string, 0)
	}
	var entry interface{}
	err := unmarshal(&entry)
	if err != nil {
		return err
	}
	switch reflect.ValueOf(entry).Kind() {
	case reflect.String:
		e.Entry = strings.Split(entry.(string), " ")
	case reflect.Slice:
		return unmarshal(&e.Entry)
	}
	return nil
}

//UnmarshalJSON customize Entrypoint unmarshal
func (e *Entrypoint) UnmarshalJSON(b []byte) error {
	if e.Entry == nil {
		e.Entry = make([]string, 0)
	}
	var entry interface{}
	err := json.Unmarshal(b, &entry)
	if err != nil {
		return err
	}
	switch reflect.ValueOf(entry).Kind() {
	case reflect.String:
		e.Entry = strings.Split(entry.(string), " ")
	case reflect.Slice:
		return json.Unmarshal(b, &e.Entry)
	}
	return nil
}

func (e Entrypoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Entry)
}

// VolumeInfo storage volume configuration
type VolumeInfo struct {
	// specifies a unique name for the storage volume
	Name string `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9][a-zA-Z0-9_-]{0\\,63}$"`
	// specifies the directory where the storage volume is on the host
	Path string `yaml:"path" json:"path" validate:"nonzero"`
	// specifies the metadata of the storage volume
	Meta Meta `yaml:"meta" json:"meta"`
}

type Meta struct {
	URL     string `yaml:"url" json:"url"`
	MD5     string `yaml:"md5" json:"md5"`
	Version string `yaml:"version" json:"version"`
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
