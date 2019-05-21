package master

import (
	"fmt"
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

func (m *Master) prepareServices() ([]openedge.VolumeInfo, []openedge.ServiceInfo, []openedge.ServiceInfo, error) {
	oldVolumes := m.appcfg.Volumes
	oldServices := m.appcfg.Services

	err := m.load()
	if err != nil {
		return nil, nil, nil, err
	}

	newVolumes := m.appcfg.Volumes
	newServices := m.appcfg.Services

	updatedVolumes, removedVolumes := openedge.DiffVolumes(oldVolumes, newVolumes)
	updatedServices, removedServices := openedge.DiffServices(oldServices, newServices, updatedVolumes)

	m.engine.Prepare(updatedServices)
	return removedVolumes, updatedServices, removedServices, nil
}

func (m *Master) startAllServices(updatedServices []openedge.ServiceInfo) error {
	services := updatedServices
	if updatedServices == nil {
		services = m.appcfg.Services
	} else if len(services) == 0 {
		return nil
	}

	vs := make(map[string]openedge.VolumeInfo)
	for _, v := range m.appcfg.Volumes {
		vs[v.Name] = v
	}
	for _, s := range services {
		cur, ok := m.services.Get(s.Name)
		if ok {
			cur.(engine.Service).Stop()
		}
		token := uuid.Generate().String()
		m.accounts.Set(s.Name, token)
		s.Env[openedge.EnvServiceNameKey] = s.Name
		s.Env[openedge.EnvServiceTokenKey] = token
		nxt, err := m.engine.Run(s, vs)
		if err != nil {
			m.log.Infof("failed to start service (%s)", s.Name)
			return err
		}
		m.services.Set(s.Name, nxt)
		m.log.Infof("service (%s) started", s.Name)
	}
	return nil
}

func (m *Master) stopAllServices(updatedServices []openedge.ServiceInfo) {
	services := updatedServices
	if updatedServices == nil {
		services = m.appcfg.Services
	} else if len(services) == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, s := range services {
		service, ok := m.services.Get(s.Name)
		if ok {
			wg.Add(1)
			go func(s engine.Service) {
				defer wg.Done()
				s.Stop()
				m.services.Remove(s.Name())
				m.accounts.Remove(s.Name())
				m.log.Infof("service (%s) stopped", s.Name())
			}(service.(engine.Service))
		}
	}

	wg.Wait()
}

// ReportInstance reports the stats of the instance of the service
func (m *Master) ReportInstance(serviceName, instanceName string, partialStats engine.PartialStats) error {
	_, ok := m.services.Get(serviceName)
	if !ok {
		return fmt.Errorf("service (%s) not found", serviceName)
	}
	m.infostats.AddInstanceStats(serviceName, instanceName, partialStats)
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
