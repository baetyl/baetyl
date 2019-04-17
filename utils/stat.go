package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

// DiskInfo disk information
type DiskInfo struct {
	Path              string  `json:"path,omitempty"`
	Fstype            string  `json:"fstype,omitempty"`
	Total             uint64  `json:"total,omitempty"`
	Free              uint64  `json:"free,omitempty"`
	Used              uint64  `json:"used,omitempty"`
	UsedPercent       float64 `json:"used_percent,omitempty"`
	InodesTotal       uint64  `json:"inodes_total,omitempty"`
	InodesUsed        uint64  `json:"inodes_used,omitempty"`
	InodesFree        uint64  `json:"inodes_free,omitempty"`
	InodesUsedPercent float64 `json:"inodes_used_percent,omitempty"`
}

// GetDiskInfo gets disk information
func GetDiskInfo(path string) (*DiskInfo, error) {
	d, err := disk.Usage(path)
	if err != nil {
		return nil, err
	}
	return &DiskInfo{
		Path:              d.Path,
		Fstype:            d.Fstype,
		Total:             d.Total,
		Free:              d.Free,
		Used:              d.Used,
		UsedPercent:       d.UsedPercent,
		InodesTotal:       d.InodesTotal,
		InodesUsed:        d.InodesUsed,
		InodesFree:        d.InodesFree,
		InodesUsedPercent: d.InodesUsedPercent,
	}, nil
}

// MemInfo memory information
type MemInfo struct {
	Total           uint64  `json:"total,omitempty"`
	Free            uint64  `json:"free,omitempty"`
	Used            uint64  `json:"used,omitempty"`
	UsedPercent     float64 `json:"used_percent,omitempty"`
	SwapTotal       uint64  `json:"swap_total,omitempty"`
	SwapFree        uint64  `json:"swap_free,omitempty"`
	SwapUsed        uint64  `json:"swap_used,omitempty"`
	SwapUsedPercent float64 `json:"swap_used_percent,omitempty"`
}

// GetMemInfo gets memory information
func GetMemInfo() (*MemInfo, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	sm, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}
	return &MemInfo{
		Total:           vm.Total,
		Free:            vm.Free,
		Used:            vm.Used,
		UsedPercent:     vm.UsedPercent,
		SwapTotal:       sm.Total,
		SwapFree:        sm.Free,
		SwapUsed:        sm.Used,
		SwapUsedPercent: sm.UsedPercent,
	}, nil
}

// CPUInfo CPU information
type CPUInfo struct {
	UsedPercent float64        `json:"used_percent,omitempty"`
	CPUs        []cpu.InfoStat `json:"cpus,omitempty"`
}

// GetCPUInfo gets CPU information
func GetCPUInfo() (*CPUInfo, error) {
	pci, err := cpu.Info()
	if err != nil {
		return nil, err
	}
	cp, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	ci := &CPUInfo{
		CPUs: pci,
	}
	if len(cp) == 1 {
		ci.UsedPercent = cp[0]
	}
	return ci, nil
}

// GPUInfo GPU information
type GPUInfo struct {
	Index          string  `json:"index,omitempty"`
	Model          string  `json:"model,omitempty"`
	MemTotal       int64   `json:"mem_total,omitempty"`
	MemFree        int64   `json:"mem_free,omitempty"`
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
func GetGPUInfo() ([]GPUInfo, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(nvSmiBin, nvQueryArg, nvFormatArg)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}
	return parseGPUInfo(stdout.String())
}

func parseGPUInfo(in string) (gpus []GPUInfo, err error) {
	for _, raw := range strings.Split(in, "\n") {
		var g GPUInfo
		parts := strings.Split(raw, ",")
		if len(parts) != 6 {
			continue
		}
		g.Index = strings.TrimSpace(parts[0])
		g.Model = strings.TrimSpace(parts[1])
		g.MemTotal, err = strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		if err != nil {
			return
		}
		g.MemFree, err = strconv.ParseInt(strings.TrimSpace(parts[3]), 10, 64)
		if err != nil {
			return
		}
		g.MemUsedPercent, err = strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)
		if err != nil {
			return
		}
		g.GPUUsedPercent, err = strconv.ParseFloat(strings.TrimSpace(parts[5]), 64)
		if err != nil {
			return
		}
		gpus = append(gpus, g)
	}
	return
}
