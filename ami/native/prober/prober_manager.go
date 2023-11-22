package prober

import (
	"strings"
	"sync"
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/kardianos/service"
	"github.com/timshannon/bolthold"
	"k8s.io/utils/clock"

	"github.com/baetyl/baetyl/v2/utils"
)

type Manager interface {
	AddApp(svc service.Service, app *v1.Application)
	RemoveApp(info *v1.AppInfo)
	CheckAndStart(svc service.Service, info *v1.AppInfo)
	CleanupApps(apps map[string]bool)
}

type manager struct {
	workers    map[probeKey]*worker
	workerLock sync.RWMutex
	start      time.Time
	log        *log.Logger
	// prober executes the probe actions.
	prober *prober
	store  *bolthold.Store
	// count when collecting process status, if count >=maxProbeRetries, stop worker
	status map[probeKey]int
}

func NewManager(store *bolthold.Store) Manager {
	return &manager{
		workers: make(map[probeKey]*worker),
		start:   clock.RealClock{}.Now(),
		prober:  newProber(),
		store:   store,
		status:  make(map[probeKey]int),
		log:     log.With(log.Any("native", "probe")),
	}
}

func (m *manager) AddApp(svc service.Service, app *v1.Application) {
	if app == nil || len(app.Services) == 0 {
		return
	}
	var p probeType
	if app.Services[0].LivenessProbe != nil {
		p = liveness
	} else if app.Services[0].StartupProbe != nil {
		p = startup
	} else {
		return
	}
	m.workerLock.Lock()
	defer m.workerLock.Unlock()
	key := probeKey{Name: app.Name, Version: app.Version, ProbeType: p}
	if _, ok := m.workers[key]; ok {
		return
	}
	w := newWorker(m, svc, p, app)
	m.workers[key] = w
	m.status[key] = 0
	m.log.Debug("add app", log.Any("app", key))
	go w.run()
}

func (m *manager) RemoveApp(app *v1.AppInfo) {
	if app == nil {
		return
	}
	m.workerLock.Lock()
	defer m.workerLock.Unlock()
	livenessKey := probeKey{Name: app.Name, Version: app.Version, ProbeType: liveness}
	if w, ok := m.workers[livenessKey]; ok {
		w.stop()
	}
	startupKey := probeKey{Name: app.Name, Version: app.Version, ProbeType: startup}
	if w, ok := m.workers[startupKey]; ok {
		w.stop()
	}
}

// CheckAndStart This is used for restarting baetyl-core, due to applying apps will not be called.
func (m *manager) CheckAndStart(svc service.Service, info *v1.AppInfo) {
	if strings.HasPrefix(info.Name, v1.BaetylCore) || strings.HasPrefix(info.Name, v1.BaetylInit) {
		return
	}
	key := utils.MakeKey(v1.KindApplication, info.Name, info.Version)
	app := new(v1.Application)
	err := m.store.Get(key, app)
	if err != nil {
		m.log.Error("failed to get app", log.Any("app", key), log.Error(err))
		return
	}
	m.AddApp(svc, app)
}

// CleanupApps removes the workers when collecting the process status.
// Only if count >=maxProbeRetries, stop worker
func (m *manager) CleanupApps(apps map[string]bool) {
	m.workerLock.Lock()
	defer m.workerLock.Unlock()
	for key, w := range m.workers {
		k := utils.MakeKey(v1.KindApplication, key.Name, key.Version)
		if _, ok := apps[k]; !ok {
			m.status[key]++
			if m.status[key] >= maxProbeRetries {
				m.log.Debug("remove app", log.Any("key", key), log.Any("apps", apps))
				w.stop()
			}
		} else {
			// reset count
			m.status[key] = 0
		}
	}
}

// Called by the worker after exiting.
func (m *manager) removeWorker(name probeKey) {
	m.workerLock.Lock()
	defer m.workerLock.Unlock()
	delete(m.workers, name)
	delete(m.status, name)
}
