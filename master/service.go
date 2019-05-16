package master

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/baidu/openedge/master/engine"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	"github.com/docker/distribution/uuid"
	cmap "github.com/orcaman/concurrent-map"
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

func (m *Master) prepareServices() ([]openedge.VolumeInfo, map[string][]openedge.ServiceInfo, error) {
	ovs := m.appcfg.Volumes
	err := m.load()
	if err != nil {
		return nil, nil, err
	}
	m.engine.Prepare(m.appcfg.Services)
	nvs := m.appcfg.Volumes
	updatedServices, err := m.getUpdatedServices()
	if err != nil {
		return nil, nil, err
	}
	return openedge.GetRemovedVolumes(ovs, nvs), updatedServices, nil
}

func (m *Master) startAllServices(updatedServices map[string][]openedge.ServiceInfo) error {
	if err := m.load(); err != nil {
		return err
	}
	services := m.appcfg.Services
	if updatedServices != nil {
		updated, ok := updatedServices["updated"]
		if ok && len(updated) != 0 {
			services = updated
		}
	}
	vs := make(map[string]openedge.VolumeInfo)
	// TODO
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

func (m *Master) stopAllServices(updatedServices map[string][]openedge.ServiceInfo) {
	target := m.services
	if updatedServices != nil {
		removed, ok := updatedServices["removed"]
		if ok && len(removed) != 0 {
			target = cmap.New()
			var serviceName string
			for _, removedService := range removed {
				serviceName = removedService.Name
				if m.services.Has(serviceName) {
					service, ok := m.services.Get(serviceName)
					if ok {
						target.Set(serviceName, service)
					}
				}
			}
		}
	}

	var wg sync.WaitGroup
	for _, s := range target.Items() {
		wg.Add(1)
		go func(s engine.Service) {
			defer wg.Done()
			s.Stop()
			m.services.Remove(s.Name())
			m.accounts.Remove(s.Name())
			m.log.Infof("service (%s) stopped", s.Name())
		}(s.(engine.Service))
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

// TODO: compare volume
func (m *Master) getUpdatedServices() (map[string][]openedge.ServiceInfo, error) {
	oldCfg := m.appcfg
	var newCfg openedge.AppConfig
	err := utils.LoadYAML(appConfigFile, &newCfg)
	if err != nil {
		return nil, err
	}

	oldServicesInfo := make(map[string]openedge.ServiceInfo)
	newServicesInfo := make(map[string]openedge.ServiceInfo)
	removed := make([]openedge.ServiceInfo, 5)
	updated := make([]openedge.ServiceInfo, 5)

	// new services info
	for _, service := range newCfg.Services {
		newServicesInfo[service.Image] = service
	}

	// old services info and removed services
	for _, service := range oldCfg.Services {
		oldServicesInfo[service.Image] = service
		_, ok := newServicesInfo[service.Image]
		if !ok {
			removed = append(removed, service)
		}
	}

	// new services and updated services
	for imageName, service := range newServicesInfo {
		oldService, ok := oldServicesInfo[imageName]
		if ok {
			if !reflect.DeepEqual(service, oldService) {
				removed = append(removed, oldService)
				updated = append(updated, service)
			}
		} else {
			updated = append(updated, service)
		}
	}

	updatedServices := map[string][]openedge.ServiceInfo{
		"updated": updated,
		"removed": removed,
	}

	return updatedServices, nil
}
