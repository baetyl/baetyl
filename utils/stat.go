package utils

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

// HostInfo host information
type HostInfo struct {
	Time            time.Time `json:"time,omitempty"`
	Hostname        string    `json:"hostname,omitempty"`
	Uptime          uint64    `json:"uptime,omitempty"`
	BootTime        uint64    `json:"boot_time,omitempty"`
	ProcessNum      uint64    `json:"process_num,omitempty"`
	OS              string    `json:"os,omitempty"`
	Platform        string    `json:"platform,omitempty"`
	PlatformFamily  string    `json:"platform_family,omitempty"`
	PlatformVersion string    `json:"platform_version,omitempty"`
	KernelVersion   string    `json:"kernel_version,omitempty"`
	HostID          string    `json:"host_id,omitempty"`
	Error           string    `json:"error,omitempty"`
}

// GetHostInfo returns host information
func GetHostInfo() *HostInfo {
	hi := &HostInfo{Time: time.Now().UTC()}
	raw, err := host.Info()
	if err != nil {
		hi.Error = err.Error()
		return hi
	}
	hi.Hostname = raw.Hostname
	hi.Uptime = raw.Uptime
	hi.BootTime = raw.BootTime
	hi.ProcessNum = raw.Procs
	hi.OS = raw.OS
	hi.Platform = raw.Platform
	hi.PlatformFamily = raw.PlatformFamily
	hi.PlatformVersion = raw.PlatformVersion
	hi.KernelVersion = raw.KernelVersion
	hi.HostID = raw.HostID
	return hi
}

// DiskInfo disk information
type DiskInfo struct {
	Time              time.Time `json:"time,omitempty"`
	Path              string    `json:"path,omitempty"`
	Fstype            string    `json:"fstype,omitempty"`
	Total             uint64    `json:"total,omitempty"`
	Free              uint64    `json:"free,omitempty"`
	Used              uint64    `json:"used,omitempty"`
	UsedPercent       float64   `json:"used_percent,omitempty"`
	InodesTotal       uint64    `json:"inodes_total,omitempty"`
	InodesUsed        uint64    `json:"inodes_used,omitempty"`
	InodesFree        uint64    `json:"inodes_free,omitempty"`
	InodesUsedPercent float64   `json:"inodes_used_percent,omitempty"`
	Error             string    `json:"error,omitempty"`
}

// GetDiskInfo gets disk information
func GetDiskInfo(path string) *DiskInfo {
	di := &DiskInfo{Time: time.Now().UTC()}
	raw, err := disk.Usage(path)
	if err != nil {
		di.Error = err.Error()
		return di
	}
	di.Path = raw.Path
	di.Fstype = raw.Fstype
	di.Total = raw.Total
	di.Free = raw.Free
	di.Used = raw.Used
	di.UsedPercent = raw.UsedPercent
	di.InodesTotal = raw.InodesTotal
	di.InodesUsed = raw.InodesUsed
	di.InodesFree = raw.InodesFree
	di.InodesUsedPercent = raw.InodesUsedPercent
	return di
}

// MemInfo memory information
type MemInfo struct {
	Time            time.Time `json:"time,omitempty"`
	Total           uint64    `json:"total,omitempty"`
	Free            uint64    `json:"free,omitempty"`
	Used            uint64    `json:"used,omitempty"`
	UsedPercent     float64   `json:"used_percent,omitempty"`
	SwapTotal       uint64    `json:"swap_total,omitempty"`
	SwapFree        uint64    `json:"swap_free,omitempty"`
	SwapUsed        uint64    `json:"swap_used,omitempty"`
	SwapUsedPercent float64   `json:"swap_used_percent,omitempty"`
	Error           string    `json:"error,omitempty"`
}

// GetMemInfo gets memory information
func GetMemInfo() *MemInfo {
	mi := &MemInfo{Time: time.Now().UTC()}
	vm, err := mem.VirtualMemory()
	if err != nil {
		mi.Error = err.Error()
		return mi
	}
	mi.Total = vm.Total
	mi.Free = vm.Free
	mi.Used = vm.Used
	mi.UsedPercent = vm.UsedPercent
	sm, err := mem.SwapMemory()
	if err != nil {
		mi.Error = err.Error()
		return mi
	}
	mi.SwapTotal = sm.Total
	mi.SwapFree = sm.Free
	mi.SwapUsed = sm.Used
	mi.SwapUsedPercent = sm.UsedPercent
	return mi
}

// NetInfo host information
type NetInfo struct {
	Time       time.Time   `json:"time,omitempty"`
	Interfaces []Interface `json:"interfaces,omitempty"`
	Error      string      `json:"error,omitempty"`
}

// Interface interface information
type Interface struct {
	Index int    `json:"index,omitempty"`
	Name  string `json:"name,omitempty"`
	MAC   string `json:"mac,omitempty"`
	Addrs []Addr `json:"addrs,omitempty"`
	Error string `json:"error,omitempty"`
}

// Addr network ip address
type Addr struct {
	Network string `json:"network,omitempty"`
	Address string `json:"address,omitempty"`
}

// GetNetInfo returns host information
func GetNetInfo() *NetInfo {
	ni := &NetInfo{Time: time.Now().UTC(), Interfaces: []Interface{}}
	raw, err := net.Interfaces()
	if err != nil {
		ni.Error = err.Error()
		return ni
	}
	for _, v := range raw {
		i := Interface{
			Index: v.Index,
			Name:  v.Name,
			MAC:   v.HardwareAddr.String(),
			Addrs: []Addr{},
		}
		va, err := v.Addrs()
		if err != nil {
			i.Error = err.Error()
		}
		for _, vaa := range va {
			i.Addrs = append(i.Addrs, Addr{Network: vaa.Network(), Address: vaa.String()})
		}
		ni.Interfaces = append(ni.Interfaces, i)
	}
	return ni
}

// CPUInfo CPU information
type CPUInfo struct {
	Time        time.Time `json:"time,omitempty"`
	Mhz         float64   `json:"mhz,omitempty"`
	Cores       int32     `json:"cores,omitempty"`
	CacheSize   int32     `json:"cache_size,omitempty"`
	ModelName   string    `json:"model_name,omitempty"`
	PhysicalID  string    `json:"physical_id,omitempty"`
	UsedPercent float64   `json:"used_percent,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// PerCPUInfo one CPU information
type PerCPUInfo struct {
	UsedPercent float64 `json:"used_percent,omitempty"`
}

// GetCPUInfo gets CPU information
func GetCPUInfo() *CPUInfo {
	ci := &CPUInfo{Time: time.Now().UTC()}
	info, err := cpu.Info()
	if err != nil {
		ci.Error = err.Error()
		return ci
	}
	raw, err := cpu.Percent(0, false)
	if err != nil {
		ci.Error = err.Error()
		return ci
	}
	// only get first CPU
	if len(info) >= 1 {
		ci.Mhz = info[0].Mhz
		ci.Cores = info[0].Cores
		ci.CacheSize = info[0].CacheSize
		ci.ModelName = info[0].ModelName
		ci.PhysicalID = info[0].PhysicalID
	}
	if len(raw) >= 1 {
		ci.UsedPercent = raw[0]
	}
	return ci
}

// GPUInfo GPU information
type GPUInfo struct {
	Time  time.Time    `json:"time,omitempty"`
	GPUs  []PerGPUInfo `json:"gpus,omitempty"`
	Error string       `json:"error,omitempty"`
}

// PerGPUInfo one GPU information
type PerGPUInfo struct {
	Index          string  `json:"index,omitempty"`
	Model          string  `json:"model,omitempty"`
	MemTotal       uint64  `json:"mem_total,omitempty"`
	MemFree        uint64  `json:"mem_free,omitempty"`
	MemUsedPercent float64 `json:"mem_used_percent,omitempty"`
	GPUUsedPercent float64 `json:"gpu_used_percent,omitempty"`
}

/********************************************************************************************
* nvidia-smi --query-gpu=index,name,memory.total,memory.free,utilization.memory,utilization.gpu --format=csv,noheader,nounits
* 0, TITAN X (Pascal), 12189, 12187, 0, 0
* 1, TITAN X (Pascal), 12189, 12187, 0, 0
* 2, TITAN X (Pascal), 12189, 12187, 0, 0
* 3, TITAN X (Pascal), 12189, 12187, 0, 0
********************************************************************************************/

const (
	nvSmiBin    = "nvidia-smi"
	nvQueryArg  = "--query-gpu=index,name,memory.total,memory.free,utilization.memory,utilization.gpu"
	nvFormatArg = "--format=csv,noheader,nounits"
)

// GetGPUInfo gets GPU information
func GetGPUInfo() *GPUInfo {
	var stderr, stdout bytes.Buffer
	gi := &GPUInfo{Time: time.Now().UTC()}
	cmd := exec.Command(nvSmiBin, nvQueryArg, nvFormatArg)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		gi.Error = fmt.Sprintf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	} else {
		gi.GPUs = parseGPUInfo(stdout.String())
	}
	return gi
}

func parseGPUInfo(in string) []PerGPUInfo {
	var err error
	gpus := []PerGPUInfo{}
	for _, raw := range strings.Split(in, "\n") {
		var g PerGPUInfo
		parts := strings.Split(raw, ",")
		if len(parts) != 6 {
			continue
		}
		g.Index = strings.TrimSpace(parts[0])
		g.Model = strings.TrimSpace(parts[1])
		g.MemTotal, err = strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64)
		if err != nil {
			continue
		}
		g.MemFree, err = strconv.ParseUint(strings.TrimSpace(parts[3]), 10, 64)
		if err != nil {
			continue
		}
		g.MemUsedPercent, err = strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)
		if err != nil {
			continue
		}
		g.GPUUsedPercent, err = strconv.ParseFloat(strings.TrimSpace(parts[5]), 64)
		if err != nil {
			continue
		}
		gpus = append(gpus, g)
	}
	return gpus
}
