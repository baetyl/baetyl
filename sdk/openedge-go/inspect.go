package openedge

import (
	"time"

	"github.com/baidu/openedge/utils"
)

// Inspect openedge information inspected
type Inspect struct {
	Error    string    `json:"error,omitempty"`
	Time     time.Time `json:"time,omitempty"`
	Software Software  `json:"software,omitempty"`
	Hardware Hardware  `json:"hardware,omitempty"`
	Services Services  `json:"services,omitempty"`
	// Volumes  []VolumeStatus `json:"volumes,omitempty"`
}

// Software software information
type Software struct {
	OS          string `json:"os,omitempty"`
	Arch        string `json:"arch,omitempty"`
	Mode        string `json:"mode,omitempty"`
	GoVersion   string `json:"go_version,omitempty"`
	BinVersion  string `json:"bin_version,omitempty"`
	ConfVersion string `json:"conf_version,omitempty"`
}

// Hardware hardware information
type Hardware struct {
	MemInfo  *utils.MemInfo  `json:"mem_stats,omitempty"`
	CPUInfo  *utils.CPUInfo  `json:"cpu_stats,omitempty"`
	DiskInfo *utils.DiskInfo `json:"disk_stats,omitempty"`
	GPUInfo  []utils.GPUInfo `json:"gpu_stats,omitempty"`
}

// Services all services' information
type Services []ServiceStatus

// ServiceStatus service status
type ServiceStatus struct {
	Name      string           `json:"name,omitempty"`
	Instances []InstanceStatus `json:"instances,omitempty"`
}

// InstanceStatus service instance status
type InstanceStatus map[string]interface{}

// NewInspect create a new information inspected
func NewInspect() *Inspect {
	return &Inspect{
		Services: Services{},
	}
}

// NewServiceStatus create a new service status
func NewServiceStatus(name string) ServiceStatus {
	return ServiceStatus{
		Name:      name,
		Instances: []InstanceStatus{},
	}
}
