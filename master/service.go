package master

import (
	"fmt"
	"sync"

	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/sdk-go/openedge"
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

func (m *Master) startAllServices() error {
	err := m.load()
	if err != nil {
		return err
	}
	vs := make(map[string]openedge.VolumeInfo)
	for _, v := range m.appcfg.Volumes {
		vs[v.Name] = v
	}
	for _, s := range m.appcfg.Services {
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
			return err
		}
		m.services.Set(s.Name, nxt)
	}
	return nil
}

func (m *Master) stopAllServices() {
	var wg sync.WaitGroup
	for _, s := range m.services.Items() {
		wg.Add(1)
		go func(s engine.Service) {
			defer wg.Done()
			s.Stop()
			m.services.Remove(s.Name())
			m.accounts.Remove(s.Name())
		}(s.(engine.Service))
	}
	wg.Wait()
}

// StartServiceInstance starts a service instance
func (m *Master) StartServiceInstance(service, instance string, dynamicConfig map[string]string) error {
	s, ok := m.services.Get(service)
	if !ok {
		return fmt.Errorf("service (%s) not found", service)
	}
	return s.(engine.Service).StartInstance(instance, dynamicConfig)
}

// StopServiceInstance stops a service instance
func (m *Master) StopServiceInstance(service, instance string) error {
	s, ok := m.services.Get(service)
	if !ok {
		return fmt.Errorf("service (%s) not found", service)
	}
	return s.(engine.Service).StopInstance(instance)
}
