package master

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	"gopkg.in/yaml.v2"
)

type infoStats struct {
	openedge.Inspect
	services engine.ServicesStats
	file     string
	sync.RWMutex
}

func newInfoStats(pwd, mode, version, file string) *infoStats {
	return &infoStats{
		file:     file,
		services: engine.ServicesStats{},
		Inspect: openedge.Inspect{
			Software: openedge.Software{
				OS:         runtime.GOOS,
				Arch:       runtime.GOARCH,
				GoVersion:  runtime.Version(),
				PWD:        pwd,
				Mode:       mode,
				BinVersion: version,
			},
			Hardware: openedge.Hardware{
				HostInfo: utils.GetHostInfo(),
				NetInfo:  utils.GetNetInfo(),
			},
		},
	}
}

func (is *infoStats) SetInstanceStats(serviceName, instanceName string, partialStats engine.PartialStats, persist bool) {
	is.Lock()
	service, ok := is.services[serviceName]
	if !ok {
		service = engine.InstancesStats{}
		is.services[serviceName] = service
	}
	instance, ok := service[instanceName]
	if !ok {
		instance = partialStats
		service[instanceName] = instance
	} else {
		for k, v := range partialStats {
			instance[k] = v
		}
	}
	if persist {
		is.persistStats()
	}
	is.Unlock()
}

func (is *infoStats) DelInstanceStats(serviceName, instanceName string, persist bool) {
	is.Lock()
	defer is.Unlock()
	service, ok := is.services[serviceName]
	if !ok {
		return
	}
	_, ok = service[instanceName]
	if !ok {
		return
	}
	delete(service, instanceName)
	if persist {
		is.persistStats()
	}
}

func (is *infoStats) setVersion(ver string) {
	is.Lock()
	is.Inspect.Software.ConfVersion = ver
	is.Unlock()
}

func (is *infoStats) getVersion() string {
	is.RLock()
	defer is.RUnlock()
	return is.Inspect.Software.ConfVersion
}

func (is *infoStats) setError(err error) {
	is.Lock()
	if err == nil {
		is.Inspect.Error = ""
	} else {
		is.Inspect.Error = err.Error()
	}
	is.Unlock()
}

func (is *infoStats) getError() string {
	is.RLock()
	defer is.RUnlock()
	return is.Inspect.Error
}

// func genVolumesStats(cfg []openedge.VolumeInfo) openedge.Volumes {
// 	volumes := openedge.Volumes{}
// 	for _, item := range cfg {
// 		volumes = append(volumes, openedge.VolumeStatus{
// 			Name:    item.Name,
// 			Version: item.Meta.Version,
// 		})
// 	}
// 	return volumes
// }

func (is *infoStats) persistStats() {
	data, err := yaml.Marshal(is.services)
	if err != nil {
		logger.Global.WithError(err).Warnf("failed to persist services stats")
		return
	}
	err = ioutil.WriteFile(is.file, data, 0755)
	if err != nil {
		logger.Global.WithError(err).Warnf("failed to persist services stats")
	}
}

func (is *infoStats) LoadStats(services interface{}) bool {
	if !utils.FileExists(is.file) {
		return false
	}
	data, err := ioutil.ReadFile(is.file)
	if err != nil {
		logger.Global.WithError(err).Warnf("failed to read old stats")
		os.Rename(is.file, fmt.Sprintf("%s.%d", is.file, time.Now().Unix()))
		return false
	}
	err = yaml.Unmarshal(data, services)
	if err != nil {
		logger.Global.WithError(err).Warnf("failed to unmarshal old stats")
		os.Rename(is.file, fmt.Sprintf("%s.%d", is.file, time.Now().Unix()))
		return false
	}
	return true
}

func (is *infoStats) stats() {
	t := time.Now().UTC()
	gi := utils.GetGPUInfo()
	mi := utils.GetMemInfo()
	ci := utils.GetCPUInfo()
	di := utils.GetDiskInfo("/")

	is.Lock()
	is.Inspect.Time = t
	is.Inspect.Hardware.GPUInfo = gi
	is.Inspect.Hardware.MemInfo = mi
	is.Inspect.Hardware.CPUInfo = ci
	is.Inspect.Hardware.DiskInfo = di
	is.Unlock()
}

// InspectSystem inspects info and stats of openedge system
func (m *Master) InspectSystem() *openedge.Inspect {
	defer utils.Trace("InspectSystem", logger.Global.Debugf)()
	var wg sync.WaitGroup
	for item := range m.services.IterBuffered() {
		wg.Add(1)
		go func(s engine.Service) {
			defer wg.Done()
			s.Stats()
		}(item.Val.(engine.Service))
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.infostats.stats()
	}()
	wg.Wait()

	result := m.infostats.Inspect
	result.Services = openedge.Services{}
	for serviceName, serviceStats := range m.infostats.services {
		service := openedge.NewServiceStatus(serviceName)
		for _, instanceStats := range serviceStats {
			service.Instances = append(service.Instances, map[string]interface{}(instanceStats))
		}
		result.Services = append(result.Services, service)
	}
	return &result
}
