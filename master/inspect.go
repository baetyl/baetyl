package master

import (
	"sync"
	"time"

	"github.com/baidu/openedge/master/engine"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
)

// InspectSystem inspects system
func (m *Master) InspectSystem() *openedge.Inspect {
	defer utils.Trace("InspectSystem", m.log.Debugf)()

	m.stats.Time = time.Now().UTC()
	m.stats.Software.ConfVersion = m.appcfg.Version

	var err error

	if m.stats.Hardware.HostInfo == nil {
		m.stats.Hardware.HostInfo, err = utils.GetHostInfo()
		if err != nil {
			m.log.Debugf("failed to get host information: %s", err.Error())
		}
	}
	if m.stats.Hardware.NetInfo == nil {
		m.stats.Hardware.NetInfo, err = utils.GetNetInfo()
		if err != nil {
			m.log.Debugf("failed to get net information: %s", err.Error())
		}
	}
	m.stats.Hardware.GPUInfo, err = utils.GetGPUInfo()
	if err != nil {
		m.log.Debugf("failed to get gpu information: %s", err.Error())
	}
	m.stats.Hardware.MemInfo, err = utils.GetMemInfo()
	if err != nil {
		m.log.Debugf("failed to get memory information: %s", err.Error())
	}
	m.stats.Hardware.CPUInfo, err = utils.GetCPUInfo()
	if err != nil {
		m.log.Debugf("failed to get cpu information: %s", err.Error())
	}
	m.stats.Hardware.DiskInfo, err = utils.GetDiskInfo("/")
	if err != nil {
		m.log.Debugf("failed to get disk information: %s", err.Error())
	}
	m.stats.Services = m.statServices()
	m.stats.Volumes = m.statVolumes()
	return m.stats
}

func (m *Master) statServices() openedge.Services {
	services := m.services.Items()
	results := make(chan openedge.ServiceStatus, len(services))

	var wg sync.WaitGroup
	for _, item := range services {
		wg.Add(1)
		go func(s engine.Service, wg *sync.WaitGroup) {
			defer wg.Done()
			results <- s.Stats()
		}(item.(engine.Service), &wg)
	}
	wg.Wait()
	close(results)
	r := openedge.Services{}
	for s := range results {
		r = append(r, s)
	}
	return r
}

func (m *Master) statVolumes() openedge.Volumes {
	r := openedge.Volumes{}
	for _, v := range m.appcfg.Volumes {
		r = append(r, openedge.VolumeStatus{
			Name:    v.Name,
			Version: v.Meta.Version,
		})
	}
	return r
}
