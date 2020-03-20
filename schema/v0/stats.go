package schema

import (
	"time"

	"github.com/baetyl/baetyl/utils"
)

// Stats all baetyl information and status inspected
type Stats struct {
	// exception information
	Error string `json:"error,omitempty"`
	// inspect time
	Time time.Time `json:"time,omitempty"`
	// software information
	Software SoftwareStats `json:"software,omitempty"`
	// hardware information
	Hardware HardwareStats `json:"hardware,omitempty"`
	// service information, including service name, instance running status, etc.
	Services map[string]ServiceStats `json:"services,omitempty"`
	// storage volume information, including name and version
	Volumes []VolumeStats `json:"volumes,omitempty"`
}

// ServiceStats service information
type ServiceStats struct {
	Instances map[string]InstanceStats `json:"instances,omitempty"`
}

// InstanceStats service instance information
type InstanceStats map[string]interface{}

// VolumeStats volume information
type VolumeStats struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// SoftwareStats software information
type SoftwareStats struct {
	// operating system information of host
	OS string `json:"os,omitempty"`
	// CPU information of host
	Arch string `json:"arch,omitempty"`
	// Baetyl process work directory
	PWD string `json:"pwd,omitempty"`
	// Baetyl running mode of application services
	Mode string `json:"mode,omitempty"`
	// Baetyl compiled Golang version
	GoVersion string `json:"go_version,omitempty"`
	// Baetyl release version
	BinVersion string `json:"bin_version,omitempty"`
	// Baetyl git revision
	GitRevision string `json:"git_revision,omitempty"`
	// Baetyl loaded application configuration version
	ConfVersion string `json:"conf_version,omitempty"`
}

// HardwareStats hardware information
type HardwareStats struct {
	// host information
	HostInfo *utils.HostInfo `json:"host_stats,omitempty"`
	// net information of host
	NetInfo *utils.NetInfo `json:"net_stats,omitempty"`
	// memory usage information of host
	MemInfo *utils.MemInfo `json:"mem_stats,omitempty"`
	// CPU usage information of host
	CPUInfo *CPUInfo `json:"cpu_stats,omitempty"`
	// disk usage information of host
	DiskInfo *utils.DiskInfo `json:"disk_stats,omitempty"`
	// CPU usage information of host
	GPUInfo []GPUInfo `json:"gpu_stats,omitempty"`
}

// CPUInfoV0 CPU information
type CPUInfo struct {
	UsedPercent float64 `json:"used_percent,omitempty"`
}

// GPUInfoV0 GPU information
type GPUInfo struct {
	ID    string        `json:"id,omitempty"`
	Model string        `json:"model,omitempty"`
	Mem   utils.MemInfo `json:"mem_stat,omitempty"`
}
