package schema

import (
	"time"

	v0 "github.com/baetyl/baetyl/schema/v0"
	"github.com/baetyl/baetyl/utils"
)

type SoftwareStats = v0.SoftwareStats
type ServiceStats = v0.ServiceStats
type InstanceStats = v0.InstanceStats
type VolumeStats = v0.VolumeStats

// Stats of whole baetyl
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

func (s *Stats) ToV0() *v0.Stats {
	rv := &v0.Stats{
		Time:     s.Time,
		Error:    s.Error,
		Software: v0.SoftwareStats(s.Software),
	}
	ss := make(map[string]v0.ServiceStats)
	for k, v := range s.Services {
		ss[k] = v0.ServiceStats(v)
	}
	rv.Services = ss
	vs := make([]v0.VolumeStats, 0)
	for _, v := range s.Volumes {
		vs = append(vs, v0.VolumeStats(v))
	}
	rv.Volumes = vs
	rv.Hardware.HostInfo = s.Hardware.HostInfo
	rv.Hardware.DiskInfo = s.Hardware.DiskInfo
	rv.Hardware.NetInfo = s.Hardware.NetInfo
	rv.Hardware.MemInfo = s.Hardware.MemInfo
	rv.Hardware.CPUInfo = &v0.CPUInfo{}
	rv.Hardware.CPUInfo.UsedPercent = s.Hardware.CPUInfo.UsedPercent
	gi := make([]v0.GPUInfo, 0)
	for _, v := range s.Hardware.GPUInfo.GPUs {
		gi = append(gi, v0.GPUInfo{
			ID:    v.Index,
			Model: v.Model,
			Mem: utils.MemInfo{
				Total:       v.MemTotal,
				Free:        v.MemFree,
				UsedPercent: v.MemUsedPercent,
			},
		})
	}
	rv.Hardware.GPUInfo = gi
	return rv
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
	CPUInfo *utils.CPUInfo `json:"cpu_stats,omitempty"`
	// disk usage information of host
	DiskInfo *utils.DiskInfo `json:"disk_stats,omitempty"`
	// CPU usage information of host
	GPUInfo *utils.GPUInfo `json:"gpu_stats,omitempty"`
}
