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
	// defer Trace("GetDiskInfo", logger.Debugf)()

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
	// defer Trace("GetMemInfo", logger.Debugf)()

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
	CPUs        int     `json:"cpus,omitempty"`
	UsedPercent float64 `json:"used_percent,omitempty"`
}

// GetCPUInfo gets CPU information
func GetCPUInfo() (*CPUInfo, error) {
	// defer Trace("GetCPUInfo", logger.Debugf)()

	cc, err := cpu.Counts(false)
	if err != nil {
		return nil, err
	}
	cp, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	ci := &CPUInfo{
		CPUs: cc,
	}
	if len(cp) == 1 {
		ci.UsedPercent = cp[0]
	}
	return ci, nil
}

// GPUInfo GPU information
type GPUInfo struct {
	ID    string  `json:"id,omitempty"`
	Model string  `json:"model,omitempty"`
	Mem   MemInfo `json:"mem_stat,omitempty"`
}

// GetGPUInfo gets GPU information
func GetGPUInfo() ([]GPUInfo, error) {
	// defer Trace("GetGPUInfo", logger.Debugf)()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("/bin/bash", "-c", `nvidia-smi --query-gpu=index,name,memory.total,memory.free --format=csv,noheader,nounits`)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))

	}
	var gpus []GPUInfo
	for _, raw := range strings.Split(stdout.String(), "\n") {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		var g GPUInfo
		t := strings.Split(raw, ",")
		total, err := strconv.Atoi(strings.TrimSpace(t[2]))
		if err != nil {
			return gpus, err
		}
		free, err := strconv.Atoi(strings.TrimSpace(t[3]))
		if err != nil {
			return gpus, err
		}
		g.ID = strings.TrimSpace(t[0])
		g.Model = strings.TrimSpace(t[1])
		g.Mem.Total = uint64(total * 1024 * 1024)
		g.Mem.Free = uint64(free * 1024 * 1024)
		g.Mem.Used = g.Mem.Total - g.Mem.Free
		g.Mem.UsedPercent = float64(g.Mem.Used) / float64(g.Mem.Total) * 100
		gpus = append(gpus, g)
	}
	return gpus, nil
}
