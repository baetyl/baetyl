package master

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
)

// InspectSystem inspects system
func (m *Master) InspectSystem() *openedge.Inspect {
	ms := openedge.NewInspect()
	ms.Time = time.Now().UTC()
	ms.Platform.Mode = m.cfg.Mode
	ms.Platform.GoVersion = runtime.Version()
	ms.Platform.BinVersion = Version
	ms.Platform.ConfVersion = m.appcfg.Version
	ms.HostInfo["os"] = runtime.GOOS
	ms.HostInfo["arch"] = runtime.GOARCH
	gpus, err := utils.GetGpu()
	if err != nil {
		m.log.Debugf("failed to get gpu information: %s", err.Error())
	}
	for _, gpu := range gpus {
		ms.HostInfo[fmt.Sprintf("gpu%s", gpu.ID)] = gpu.Model
		ms.HostInfo[fmt.Sprintf("gpu%s_mem_total", gpu.ID)] = gpu.Memory.Total
		ms.HostInfo[fmt.Sprintf("gpu%s_mem_free", gpu.ID)] = gpu.Memory.Free
	}
	mem, err := utils.GetMem()
	if err != nil {
		m.log.Debugf("failed to get memory information: %s", err.Error())
	}
	if mem != nil {
		ms.HostInfo["mem_total"] = mem.Total
		ms.HostInfo["mem_free"] = mem.Free
	}
	swap, err := utils.GetSwap()
	if err != nil {
		m.log.Debugf("failed to get swap information: %s", err.Error())
	}
	if swap != nil {
		ms.HostInfo["swap_total"] = swap.Total
		ms.HostInfo["swap_free"] = swap.Free
	}
	/*
		disk, err := utils.GetDisk()
		if err != nil {
			log.WithError(err).HostInfo("failed to get disk information")
		}
		if disk != nil {
			ms.HostInfo["disk_total"] = disk.Total
			ms.HostInfo["disk_free"] = disk.Free
		}
	*/
	v, ok := m.context.Get("error")
	if ok {
		ms.Error = v.(string)
	}
	ms.Services = m.statServices()
	return ms
}

func (m *Master) statServices() openedge.Services {
	services := m.services.Items()
	results := make(chan openedge.ServiceStatus, len(services))

	var wg sync.WaitGroup
	for _, s := range services {
		wg.Add(1)
		go func(s engine.Service, wg *sync.WaitGroup) {
			defer wg.Done()
			results <- s.Stats()
		}(s.(engine.Service), &wg)
	}
	wg.Wait()
	close(results)
	r := openedge.Services{}
	for s := range results {
		r = append(r, s)
	}
	return r
}
