package master

import (
	"reflect"
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/protocol/http"
	"github.com/baidu/openedge/sdk-go/openedge"
)

// Config master init config
type Config struct {
	Mode          string          `yaml:"mode" json:"mode" default:"docker" validate:"regexp=^(native|docker)$"`
	Server        http.ServerInfo `yaml:"server" json:"server"`
	DynamicConfig `yaml:",inline" json:",inline"`
}

// DynamicConfig update config
type DynamicConfig struct {
	Version  string                 `yaml:"version" json:"version"`
	Services []engine.ServiceInfo   `yaml:"services" json:"services" default:"[]"`
	Datasets []openedge.DatasetInfo `yaml:"datasets" json:"datasets" default:"[]"`
	Grace    time.Duration          `yaml:"grace" json:"grace" default:"30s"`
	Logger   logger.LogInfo         `yaml:"logger" json:"logger"`
}

type dynamicConfigDiff struct {
	startServices  []engine.ServiceInfo
	stopServices   []engine.ServiceInfo
	addDatasets    []openedge.DatasetInfo
	removeDatasets []openedge.DatasetInfo
	updateGrace    *time.Duration
	updateLogger   *logger.LogInfo
}

func (cur *DynamicConfig) diff(pre *DynamicConfig) (*dynamicConfigDiff, bool) {
	if reflect.DeepEqual(cur.Services, pre.Services) {
		return nil, true
	}

	d := new(dynamicConfigDiff)
	d.startServices = []engine.ServiceInfo{}
	d.stopServices = []engine.ServiceInfo{}
	d.addDatasets = []openedge.DatasetInfo{}
	d.removeDatasets = []openedge.DatasetInfo{}

	if !reflect.DeepEqual(cur.Services, pre.Services) {
		curServices := map[string]engine.ServiceInfo{}
		preServices := map[string]engine.ServiceInfo{}
		for _, s := range cur.Services {
			curServices[s.Name] = s
		}
		for _, s := range pre.Services {
			preServices[s.Name] = s
		}
		for n, c := range curServices {
			p, ok := preServices[n]
			if !ok {
				d.startServices = append(d.startServices, c)
				continue
			}
			if !reflect.DeepEqual(c, p) {
				d.stopServices = append(d.stopServices, p)
				d.startServices = append(d.startServices, c)
			}
		}
		for n, p := range preServices {
			_, ok := curServices[n]
			if !ok {
				d.stopServices = append(d.stopServices, p)
			}
		}
	}

	if !reflect.DeepEqual(cur.Datasets, pre.Datasets) {
		curDatasets := map[string]openedge.DatasetInfo{}
		preDatasets := map[string]openedge.DatasetInfo{}
		for _, d := range cur.Datasets {
			curDatasets[d.Name+"`"+d.Version] = d
		}
		for _, d := range pre.Datasets {
			preDatasets[d.Name+"`"+d.Version] = d
		}
		for n, c := range curDatasets {
			_, ok := preDatasets[n]
			if !ok {
				d.addDatasets = append(d.addDatasets, c)
			}
		}
		for n, p := range preDatasets {
			_, ok := curDatasets[n]
			if !ok {
				d.removeDatasets = append(d.removeDatasets, p)
			}
		}
	}

	if !reflect.DeepEqual(cur.Grace, pre.Grace) {
		d.updateGrace = &cur.Grace
	}

	if !reflect.DeepEqual(cur.Logger, pre.Logger) {
		d.updateLogger = &cur.Logger
	}
	return d, false
}
