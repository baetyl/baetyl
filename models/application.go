package models

import (
	"github.com/docker/go-units"
	"time"
)

type Application struct {
	Name       string            `json:"name,omitempty" validate:"omitempty,resourceName"`
	Namespace  string            `json:"namespace,omitempty"`
	Version    string            `json:"version,omitempty"`
	Annotation map[string]string `json:"annotation,omitempty"`
	Replicas   *int              `json:"replicas,omitempty"`
	Services   []Service         `json:"services,omitempty" binding:"required,dive"`
	Volumes    []Volume          `json:"volumes,omitempty"`
}

type Service struct {
	// specifies the unique name of the service
	Name string `json:"name,omitempty" binding:"required"`
	// specifies the hostname of the service
	Hostname string `json:"hostname,omitempty"`
	// specifies the image of the service, usually using the Docker image name
	Image string `json:"image,omitempty" binding:"required"`
	// specifies the number of instances started
	Replica int `json:"replica,omitempty" binding:"required"`
	// specifies the storage volumes that the service needs, map the storage volume to the directory in the container
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty"`
	// specifies the port bindings which exposed by the service, only for Docker container mode
	Ports []ContainerPort `json:"ports,omitempty"`
	// specifies the device bindings which used by the service, only for Docker container mode
	VolumeDevices []VolumeDevice `json:"devices,omitempty"`
	// specifies the startup arguments of the service program, but does not include `arg[0]`
	Args []string `json:"args,omitempty"`
	// specifies the environment variable of the service program
	Env []EnvVar `json:"env,omitempty"`
	// specifies the restart policy of the instance of the service
	Restart *RestartPolicyInfo `json:"restart,omitempty"`
	// specifies resource limits for a single instance of the service,  only for Docker container mode
	Resources *Resources `json:"resources,omitempty"`
	// specifies runtime to use, only for Docker container mode
	Runtime string `json:"runtime,omitempty"`
}

type Volume struct {
	// specified name of the volume
	Name string `json:"name,omitempty" binding:"required"`
	// specified store for the storage volume
	VolumeSource `json:",inline"`
}

// VolumeSource volume source, include empty directory, host path, configmap
type VolumeSource struct {
	HostPath              *HostPathVolumeSource              `json:"hostPath,omitempty"`
	Configuration         *ConfigurationVolumeSource         `json:"configuration,omitempty"`
	Secret                *SecretVolumeSource                `json:"secret,omitempty"`
	PersistentVolumeClaim *PersistentVolumeClaimVolumeSource `json:"persistentVolumeClaim"`
}

type HostPathVolumeSource struct {
	Path string `json:"path,omitempty"`
}

type ConfigurationVolumeSource struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type SecretVolumeSource struct {
	Name string `json:"name,omitempty"`
}

type PersistentVolumeClaimVolumeSource struct {
	Name     string `json:"name,omitempty"`
	ReadOnly bool   `json:"readOnly,omitempty"`
}

type VolumeMount struct {
	// specifies name of volume
	Name string `json:"name,omitempty"`
	// specifies mount path of volume
	MountPath string `json:"mountPath,omitempty"`
	// specifies if the volume is read-only
	ReadOnly bool `json:"readOnly,omitempty"`
}

// ContainerPort port map configuration
type ContainerPort struct {
	HostPort      int32  `json:"hostPort,omitempty"`
	ContainerPort int32  `json:"containerPort,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	HostIP        string `json:"hostIP,omitempty"`
}

type EnvVar struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type VolumeDevice struct {
	DevicePath string `json:"devicePath,omitempty"`
}

// RestartPolicyInfo holds the policy of a module
type RestartPolicyInfo struct {
	Retry struct {
		Max int `yaml:"max" json:"max"`
	} `yaml:"retry" json:"retry"`
	Policy string `yaml:"policy" json:"policy" default:"always"`
	//Backoff BackoffInfo `yaml:"backoff" json:"backoff"`
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
