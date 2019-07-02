package master

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/baidu/openedge/master/engine"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/docker/distribution/uuid"
)

// Auth auth api request from services
func (m *Master) Auth(username, password string) bool {
	v, ok := m.accounts.Get(username)
	if !ok {
		return false
	}
	p, ok := v.(string)
	return ok && p == password
}

func (m *Master) startServices(cur openedge.AppConfig) error {
	volumes := map[string]openedge.VolumeInfo{}
	for _, v := range cur.Volumes {
		volumes[v.Name] = v
	}
	for _, s := range cur.Services {
		if _, ok := m.services.Get(s.Name); ok {
			continue
		}
		token := uuid.Generate().String()
		m.accounts.Set(s.Name, token)
		s.Env[openedge.EnvServiceNameKey] = s.Name
		s.Env[openedge.EnvServiceTokenKey] = token
		nxt, err := m.engine.Run(s, volumes)
		if err != nil {
			m.log.Infof("failed to start service (%s)", s.Name)
			return err
		}
		m.services.Set(s.Name, nxt)
		m.log.Infof("service (%s) started", s.Name)
	}
	return nil
}

func (m *Master) stopServices(keepServices map[string]struct{}) {
	var wg sync.WaitGroup
	for item := range m.services.IterBuffered() {
		s := item.Val.(engine.Service)
		// skip the service not changed
		if _, ok := keepServices[s.Name()]; ok {
			continue
		}
		service, ok := m.services.Get(s.Name())
		if !ok {
			continue
		}
		wg.Add(1)
		go func(s engine.Service) {
			defer wg.Done()
			s.Stop()
			m.services.Remove(s.Name())
			m.accounts.Remove(s.Name())
			m.engine.DelServiceStats(s.Name(), true)
			m.log.Infof("service (%s) stopped", s.Name())
		}(service.(engine.Service))
	}
	wg.Wait()
}

// ReportInstance reports the stats of the instance of the service
func (m *Master) ReportInstance(serviceName, instanceName string, partialStats engine.PartialStats) error {
	_, ok := m.services.Get(serviceName)
	if !ok {
		return fmt.Errorf("service (%s) not found", serviceName)
	}
	m.infostats.SetInstanceStats(serviceName, instanceName, partialStats, false)
	return nil
}

// StartInstance starts a service instance
func (m *Master) StartInstance(service, instance string, dynamicConfig map[string]string) error {
	s, ok := m.services.Get(service)
	if !ok {
		return fmt.Errorf("service (%s) not found", service)
	}
	return s.(engine.Service).StartInstance(instance, dynamicConfig)
}

// StopInstance stops a service instance
func (m *Master) StopInstance(service, instance string) error {
	s, ok := m.services.Get(service)
	if !ok {
		return fmt.Errorf("service (%s) not found", service)
	}
	return s.(engine.Service).StopInstance(instance)
}

// DiffServices returns the services not changed
func diffServices(cur, old openedge.AppConfig) map[string]struct{} {
	oldVolumes := make(map[string]openedge.VolumeInfo)
	for _, o := range old.Volumes {
		oldVolumes[o.Name] = o
	}
	// find the volumes updated
	updateVolumes := make(map[string]struct{})
	for _, c := range cur.Volumes {
		if o, ok := oldVolumes[c.Name]; ok && o.Path != c.Path {
			updateVolumes[c.Name] = struct{}{}
		}
	}

	oldServices := make(map[string]openedge.ServiceInfo)
	for _, o := range old.Services {
		oldServices[o.Name] = o
	}

	// find the services not changed
	keepServices := map[string]struct{}{}
	for _, c := range cur.Services {
		o, ok := oldServices[c.Name]
		if !ok {
			continue
		}
		if !reflect.DeepEqual(c, o) {
			continue
		}
		changed := false
		for _, m := range c.Mounts {
			if _, changed = updateVolumes[m.Name]; changed {
				break
			}
		}
		if changed {
			continue
		}
		keepServices[c.Name] = struct{}{}
	}
	return keepServices
}
