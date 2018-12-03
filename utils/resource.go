package utils

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cloudfoundry/gosigar"
	"github.com/juju/errors"
)

// Disk Disk
type Disk struct {
	Total string
	Free  string
}

// GetDisk gets disk information
func GetDisk() (*Disk, error) {
	t := uint64(0)
	f := uint64(0)

	fslist := sigar.FileSystemList{}
	err := fslist.Get()
	if err != nil {
		return nil, errors.Trace(err)
	}
	for _, fs := range fslist.List {
		usage := sigar.FileSystemUsage{}
		err := usage.Get(fs.DirName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		t = t + usage.Total
		f = f + usage.Avail
	}
	disk := &Disk{}
	disk.Total = sigar.FormatSize(t * 1024)
	disk.Free = sigar.FormatSize(f * 1024)
	return disk, nil
}

// Mem Mem
type Mem struct {
	Total string
	Free  string
}

// GetMem gets memory information
func GetMem() (*Mem, error) {
	m := sigar.Mem{}
	err := m.Get()
	if err != nil {
		return nil, errors.Trace(err)
	}
	mem := &Mem{}
	mem.Total = sigar.FormatSize(m.Total)
	mem.Free = sigar.FormatSize(m.Free)
	return mem, nil
}

// Swap Swap
type Swap struct {
	Total string
	Free  string
}

// GetSwap gets swap space information
func GetSwap() (*Swap, error) {
	s := sigar.Swap{}
	err := s.Get()
	if err != nil {
		return nil, errors.Trace(err)
	}
	swap := &Swap{}
	swap.Total = sigar.FormatSize(s.Total)
	swap.Free = sigar.FormatSize(s.Free)
	return swap, nil
}

// Gpu Gpu
type Gpu struct {
	ID     string
	Model  string
	Memory Mem
}

// GetGpu gets gpu information
func GetGpu() ([]Gpu, error) {
	var gpus []Gpu
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("/bin/bash", "-c", `nvidia-smi --query-gpu=index,name,memory.total,memory.free --format=csv,noheader,nounits`)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return gpus, errors.Annotate(err, strings.Trim(stderr.String(), "\n"))

	}
	for _, raw := range strings.Split(stdout.String(), "\n") {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		var g Gpu
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
		g.Memory = Mem{
			sigar.FormatSize(uint64(total * 1024 * 1024)),
			sigar.FormatSize(uint64(free * 1024 * 1024)),
		}
		gpus = append(gpus, g)
	}
	return gpus, nil
}
