package master

import (
	"runtime"
	"sync"
	"time"

	"github.com/baidu/openedge/master/engine"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
)

// InspectSystem inspects system
func (m *Master) InspectSystem() *openedge.Inspect {
	defer utils.Trace("InspectSystem", m.log.Debugf)()

	ms := openedge.NewInspect()
	ms.Time = time.Now().UTC()
	ms.Software.OS = runtime.GOOS
	ms.Software.Arch = runtime.GOARCH
	ms.Software.Mode = m.cfg.Mode
	ms.Software.GoVersion = runtime.Version()
	ms.Software.BinVersion = Version
	ms.Software.ConfVersion = m.appcfg.Version

	var err error
	ms.Hardware.GPUInfo, err = utils.GetGPUInfo()
	if err != nil {
		m.log.Debugf("failed to get gpu information: %s", err.Error())
	}
	ms.Hardware.MemInfo, err = utils.GetMemInfo()
	if err != nil {
		m.log.Debugf("failed to get memory information: %s", err.Error())
	}
	ms.Hardware.CPUInfo, err = utils.GetCPUInfo()
	if err != nil {
		m.log.Debugf("failed to get cpu information: %s", err.Error())
	}
	ms.Hardware.DiskInfo, err = utils.GetDiskInfo("/")
	if err != nil {
		m.log.Debugf("failed to get disk information: %s", err.Error())
	}
	ms.Services = m.statServices()
	v, ok := m.context.Get("error")
	if ok {
		ms.Error = v.(string)
	}
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
