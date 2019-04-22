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
	raw  openedge.Inspect
	sss  engine.ServicesStats
	file string
	sync.RWMutex
}

func newInfoStats(pwd, mode, version, file string) *infoStats {
	return &infoStats{
		file: file,
		sss:  engine.ServicesStats{},
		raw: openedge.Inspect{
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

func (is *infoStats) AddInstanceStats(serviceName, instanceName string, partialStats engine.PartialStats) {
	is.Lock()
	service, ok := is.sss[serviceName]
	if !ok {
		service = engine.InstancesStats{}
		is.sss[serviceName] = service
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
	is.persistStats()
	is.Unlock()
}

func (is *infoStats) DelInstanceStats(serviceName, instanceName string) {
	is.Lock()
	defer is.Unlock()
	service, ok := is.sss[serviceName]
	if !ok {
		return
	}
	_, ok = service[instanceName]
	if !ok {
		return
	}
	delete(service, instanceName)
	is.persistStats()
}

func (is *infoStats) updateError(err error) {
	is.Lock()
	if err == nil {
		is.raw.Error = ""
	} else {
		is.raw.Error = err.Error()
	}
	is.Unlock()
}

func (is *infoStats) getError() string {
	is.RLock()
	defer is.RUnlock()
	return is.raw.Error
}

func (is *infoStats) refreshAppInfo(cfg openedge.AppConfig) {
	is.Lock()
	is.raw.Software.ConfVersion = cfg.Version
	is.sss = engine.ServicesStats{}
	for _, item := range cfg.Services {
		is.sss[item.Name] = engine.InstancesStats{}
	}
	is.raw.Volumes = genVolumesStats(cfg.Volumes)
	is.Unlock()
}

func genVolumesStats(cfg []openedge.VolumeInfo) openedge.Volumes {
	volumes := openedge.Volumes{}
	for _, item := range cfg {
		volumes = append(volumes, openedge.VolumeStatus{
			Name:    item.Name,
			Version: item.Meta.Version,
		})
	}
	return volumes
}

func (is *infoStats) persistStats() {
	data, err := yaml.Marshal(is.sss)
	if err != nil {
		logger.Global.WithError(err).Warnf("failed to persist services stats")
		return
	}
	err = ioutil.WriteFile(is.file, data, 0755)
	if err != nil {
		logger.Global.WithError(err).Warnf("failed to persist services stats")
	}
}

func (is *infoStats) LoadStats(sss interface{}) bool {
	if !utils.FileExists(is.file) {
		return false
	}
	data, err := ioutil.ReadFile(is.file)
	if err != nil {
		logger.Global.WithError(err).Warnf("failed to read old stats")
		os.Rename(is.file, fmt.Sprintf("%s.%d", is.file, time.Now().Unix()))
		return false
	}
	err = yaml.Unmarshal(data, sss)
	if err != nil {
		logger.Global.WithError(err).Warnf("failed to unmarshal old stats")
		os.Rename(is.file, fmt.Sprintf("%s.%d", is.file, time.Now().Unix()))
		return false
	}
	return true
}

func (is *infoStats) getInfoStats() *openedge.Inspect {
	gi := utils.GetGPUInfo()
	mi := utils.GetMemInfo()
	ci := utils.GetCPUInfo()
	di := utils.GetDiskInfo("/")

	is.Lock()
	is.raw.Hardware.GPUInfo = gi
	is.raw.Hardware.MemInfo = mi
	is.raw.Hardware.CPUInfo = ci
	is.raw.Hardware.DiskInfo = di
	result := is.raw
	is.Unlock()

	result.Services = openedge.Services{}
	for serviceName, serviceStats := range is.sss {
		service := openedge.NewServiceStatus(serviceName)
		for _, instanceStats := range serviceStats {
			service.Instances = append(service.Instances, map[string]interface{}(instanceStats))
		}
		result.Services = append(result.Services, service)
	}
	return &result
}

// InspectSystem inspects info and stats of openedge system
func (m *Master) InspectSystem() *openedge.Inspect {
	defer utils.Trace("InspectSystem", logger.Global.Debugf)()
	return m.infostats.getInfoStats()
}
