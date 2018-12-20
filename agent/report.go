package agent

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"

	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/module/logger"
	"github.com/baidu/openedge/module/utils"
)

// Report shadow's report
type Report struct {
	Reported map[string]interface{} `json:"info,omitempty"`
}

// NewReport creates a report
func NewReport(parts map[string]interface{}) *Report {
	r := &Report{Reported: parts}
	r.populateReported()
	return r
}

// Bytes converts to bytes
func (r *Report) Bytes() []byte {
	d, _ := json.Marshal(r)
	return d
}

func (r *Report) populateReported() {
	r.Reported["os"] = runtime.GOOS
	r.Reported["bit"] = strconv.IntSize
	r.Reported["arch"] = runtime.GOARCH
	r.Reported["go_version"] = runtime.Version()
	r.Reported["bin_version"] = module.Version
	gpus, err := utils.GetGpu()
	if err != nil {
		logger.WithError(err).Warnf("failed to get gpu information")
	}
	for _, gpu := range gpus {
		r.Reported[fmt.Sprintf("gpu%s", gpu.ID)] = gpu.Model
		r.Reported[fmt.Sprintf("gpu%s_mem_total", gpu.ID)] = gpu.Memory.Total
		r.Reported[fmt.Sprintf("gpu%s_mem_free", gpu.ID)] = gpu.Memory.Free
	}
	mem, err := utils.GetMem()
	if err != nil {
		logger.WithError(err).Warnf("failed to get memory information")
	}
	if mem != nil {
		r.Reported["mem_total"] = mem.Total
		r.Reported["mem_free"] = mem.Free
	}
	swap, err := utils.GetSwap()
	if err != nil {
		logger.WithError(err).Warnf("failed to get swap information")
	}
	if swap != nil {
		r.Reported["swap_total"] = swap.Total
		r.Reported["swap_free"] = swap.Free
	}
	/*
		disk, err := utils.GetDisk()
		if err != nil {
			log.WithError(err).Info("failed to get disk information")
		}
		if disk != nil {
			r.Reported["disk_total"] = disk.Total
			r.Reported["disk_free"] = disk.Free
		}
	*/
}
