package master

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/utils"
)

// Stats returns master stats
func (m *Master) stats() *engine.Stats {
	ms := engine.NewStats()
	ms.Info = make(map[string]interface{})
	ms.Info["os"] = runtime.GOOS
	ms.Info["bit"] = strconv.IntSize
	ms.Info["arch"] = runtime.GOARCH
	ms.Info["mode"] = m.cfg.Mode
	ms.Info["timestamp"] = time.Now().UTC().Unix()
	ms.Info["go_version"] = runtime.Version()
	ms.Info["bin_version"] = Version
	//ms.Info["conf_version"] = m.cfg.Version
	gpus, err := utils.GetGpu()
	if err != nil {
		openedge.Debugf("failed to get gpu information: %s", err.Error())
	}
	for _, gpu := range gpus {
		ms.Info[fmt.Sprintf("gpu%s", gpu.ID)] = gpu.Model
		ms.Info[fmt.Sprintf("gpu%s_mem_total", gpu.ID)] = gpu.Memory.Total
		ms.Info[fmt.Sprintf("gpu%s_mem_free", gpu.ID)] = gpu.Memory.Free
	}
	mem, err := utils.GetMem()
	if err != nil {
		openedge.Debugf("failed to get memory information: %s", err.Error())
	}
	if mem != nil {
		ms.Info["mem_total"] = mem.Total
		ms.Info["mem_free"] = mem.Free
	}
	swap, err := utils.GetSwap()
	if err != nil {
		openedge.Debugf("failed to get swap information: %s", err.Error())
	}
	if swap != nil {
		ms.Info["swap_total"] = swap.Total
		ms.Info["swap_free"] = swap.Free
	}
	/*
		disk, err := utils.GetDisk()
		if err != nil {
			log.WithError(err).Info("failed to get disk information")
		}
		if disk != nil {
			ms.Info["disk_total"] = disk.Total
			ms.Info["disk_free"] = disk.Free
		}
	*/
	//ms.Info["modules"] = m.engine.Stats()
	return ms
}
