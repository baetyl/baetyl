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
	updatedServices, err := m.getUpdatedServices()
	if err != nil {
		return nil, nil, err
	}
	ovs := m.appcfg.Volumes
	err = m.load()
	if err != nil {
		return nil, nil, err
	}
	m.engine.Prepare(m.appcfg.Services)
	nvs := m.appcfg.Volumes
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

func (m *Master) getUpdatedServices() (map[string][]openedge.ServiceInfo, error) {
	oldCfg := m.appcfg
	var newCfg openedge.AppConfig
	err := utils.LoadYAML(appConfigFile, &newCfg)
	if err != nil {
		return nil, err
	}

	if reflect.DeepEqual(oldCfg, newCfg) {
		return nil, nil
	}

	oldVolumes := oldCfg.Volumes
	newVolumes := newCfg.Volumes

	oldVolumesInfo := make(map[string]string)
	newVolumesInfo := make(map[string]string)

	updatedVolumesInfo := make(map[string]bool)

	for _, volume := range oldVolumes {
		newVolumesInfo[volume.Path] = volume.Name
	}

	for _, volume := range newVolumes {
		_, ok := oldVolumesInfo[volume.Path]
		if ok {
			if oldVolumesInfo[volume.Path] != volume.Name {
				updatedVolumesInfo[volume.Name] = true
			}
		} else {
			updatedVolumesInfo[volume.Name] = true
		}
	}

	oldServicesInfo := make(map[string]openedge.ServiceInfo)
	newServicesInfo := make(map[string]openedge.ServiceInfo)
	removed := make([]openedge.ServiceInfo, 0)
	updated := make([]openedge.ServiceInfo, 0)

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
	var flag bool
	for imageName, service := range newServicesInfo {
		flag = false
		oldService, ok := oldServicesInfo[imageName]
		for _, mountInfo := range service.Mounts {
			if updatedVolumesInfo[mountInfo.Name] {
				updated = append(updated, service)
				if ok {
					removed = append(removed, oldService)
				}
				flag = true
				break
			}
		}
		if flag {
			break
		}
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
