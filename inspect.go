package baetyl

import (
	gort "runtime"
	"sync"
	"time"

	"github.com/baetyl/baetyl/schema/v3"
	"github.com/baetyl/baetyl/utils"
)

type stats struct {
	sync.RWMutex
	schema.Stats
}

func newStats(mode, version, revision string) stats {
	return stats{
		Stats: schema.Stats{
			Services: map[string]schema.ServiceStats{},
			Software: schema.SoftwareStats{
				OS:          gort.GOOS,
				Arch:        gort.GOARCH,
				GoVersion:   gort.Version(),
				Mode:        mode,
				BinVersion:  version,
				GitRevision: revision,
			},
			Hardware: schema.HardwareStats{
				HostInfo: utils.GetHostInfo(),
				NetInfo:  utils.GetNetInfo(),
			},
		},
	}
}

func (rt *runtime) UpdateStats(svc, inst string, stats schema.InstanceStats) {
	rt.stats.Lock()
	s, ok := rt.stats.Services[svc]
	if !ok {
		s = schema.ServiceStats{
			Instances: map[string]schema.InstanceStats{},
		}
		rt.stats.Services[svc] = s
	}
	i, ok := s.Instances[inst]
	if !ok {
		s.Instances[inst] = stats
	} else {
		for k, v := range stats {
			i[k] = v
		}
	}
	rt.stats.Unlock()
}

func (rt *runtime) RemoveServiceStats(svc string) {
	rt.stats.Lock()
	defer rt.stats.Unlock()
	if _, ok := rt.stats.Services[svc]; ok {
		delete(rt.stats.Services, svc)
	}
}

func (rt *runtime) RemoveInstanceStats(svc, inst string) {
	rt.stats.Lock()
	defer rt.stats.Unlock()
	s, ok := rt.stats.Services[svc]
	if !ok {
		return
	}
	_, ok = s.Instances[inst]
	if !ok {
		return
	}
	delete(s.Instances, inst)
	if len(s.Instances) == 0 {
		delete(rt.stats.Services, svc)
	}
}

func (s *stats) updateHost() {
	t := time.Now().UTC()
	gi := utils.GetGPUInfo()
	mi := utils.GetMemInfo()
	ci := utils.GetCPUInfo()
	di := utils.GetDiskInfo("/")

	s.Lock()
	s.Time = t
	s.Hardware.GPUInfo = gi
	s.Hardware.MemInfo = mi
	s.Hardware.CPUInfo = ci
	s.Hardware.DiskInfo = di
	s.Unlock()
}

func (rt *runtime) inspect() *schema.Stats {
	rt.e.UpdateStats()
	rt.stats.updateHost()
	rt.log.Infoln("inspect system stats complete")
	return &rt.stats.Stats
}
