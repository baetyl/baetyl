package baetyl

// deprecated

import (
	"path"

	"github.com/baetyl/baetyl/utils"
)

// AppConfig application configuration
type AppConfig struct {
	// specifies the version of the application configuration
	Version string `yaml:"version" json:"version"`
	// specifies the service information of the application
	Services []ServiceInfo `yaml:"services" json:"services" default:"[]"`
	// specifies the storage volume information of the application
	Volumes []VolumeInfo `yaml:"volumes" json:"volumes" default:"[]"`
	// specifies the network information of the application
	Networks map[string]ComposeNetwork `yaml:"networks" json:"networks" default:"{}"`
}

// ServiceInfo service configuration
type ServiceInfo struct {
	// specifies the unique name of the service
	Name string `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9][a-zA-Z0-9_-]{0\\,63}$"`
	// specifies the hostname of the service
	Hostname string `yaml:"hostname,omitempty" json:"hostname,omitempty"`
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
	// specifies the entrypoint of the service program
	Entrypoint []string `yaml:"entrypoint" json:"entrypoint" default:"[]"`
	// specifies the environment variable of the service program
	Env map[string]string `yaml:"env" json:"env" default:"{}"`
	// specifies the restart policy of the instance of the service
	Restart RestartPolicyInfo `yaml:"restart" json:"restart"`
	// specifies resource limits for a single instance of the service,  only for Docker container mode
	Resources Resources `yaml:"resources" json:"resources"`
	// specifies runtime to use, only for Docker container mode
	Runtime string `yaml:"runtime" json:"runtime"`
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

// ToComposeAppConfig transform AppConfig into ComposeAppConfig
func (cfg AppConfig) ToComposeAppConfig() ComposeAppConfig {
	composeCfg := ComposeAppConfig{
		Version:    "3",
		AppVersion: cfg.Version,
		Networks:   cfg.Networks,
	}
	composeCfg.Volumes = map[string]ComposeVolume{}
	utils.SetDefaults(&composeCfg.Volumes)
	services := map[string]ComposeService{}
	utils.SetDefaults(&services)
	var previous string
	first := true
	for _, service := range cfg.Services {
		info := ComposeService{
			Image:       service.Image,
			Hostname:    service.Hostname,
			NetworkMode: service.NetworkMode,
			Networks:    service.Networks,
			Ports:       service.Ports,
			Devices:     service.Devices,
			Command: Command{
				Cmd: service.Args,
			},
			Entrypoint: Entrypoint{
				Entry: service.Entrypoint,
			},
			Environment: Environment{
				Envs: service.Env,
			},
			Replica:   service.Replica,
			Restart:   service.Restart,
			Resources: service.Resources,
			Runtime:   service.Runtime,
		}
		if first {
			first = false
			info.DependsOn = []string{}
		} else {
			info.DependsOn = []string{previous}
		}
		previous = service.Name
		vs := make([]ServiceVolume, 0)
		for _, mount := range service.Mounts {
			var p string
			for _, v := range cfg.Volumes {
				if v.Name == mount.Name {
					p = v.Path
				}
			}
			v := ServiceVolume{
				Source:   p,
				Target:   path.Join("/", mount.Path),
				ReadOnly: mount.ReadOnly,
			}
			vs = append(vs, v)
		}
		info.Volumes = vs
		services[service.Name] = info
	}
	composeCfg.Services = services
	return composeCfg
}
