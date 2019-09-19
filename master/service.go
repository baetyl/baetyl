package master

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/baetyl/baetyl/master/engine"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
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

func (m *Master) startServices(cur baetyl.ComposeAppConfig) error {
	for _, name := range ServiceSort(cur.Services) {
		s := cur.Services[name]
		if _, ok := m.services.Get(name); ok {
			continue
		}
		if s.ContainerName != "" {
			name = s.ContainerName
		}
		token := uuid.Generate().String()
		m.accounts.Set(name, token)
		s.Environment.Envs[baetyl.EnvKeyServiceName] = name
		s.Environment.Envs[baetyl.EnvKeyServiceToken] = token
		// TODO: remove, backward compatibility
		s.Environment.Envs[baetyl.EnvServiceNameKey] = name
		s.Environment.Envs[baetyl.EnvServiceTokenKey] = token
		nxt, err := m.engine.Run(name, s, cur.Volumes)
		if err != nil {
			m.log.Infof("failed to start service (%s)", name)
			return err
		}
		m.services.Set(name, nxt)
		m.log.Infof("service (%s) started", name)
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
func diffServices(cur, old baetyl.ComposeAppConfig) map[string]struct{} {
	// find the volumes updated
	updateVolumes := make(map[string]struct{})
	for name, c := range cur.Volumes {
		if !reflect.DeepEqual(c, old.Volumes[name]) {
			updateVolumes[name] = struct{}{}
		}
	}
	updateNetworks := make(map[string]struct{})
	for name, c := range cur.Networks {
		if !reflect.DeepEqual(c, old.Networks[name]) {
			updateNetworks[name] = struct{}{}
		}
	}
	// find the services not changed
	keepServices := map[string]struct{}{}
	for name, c := range cur.Services {
		o, ok := old.Services[name]
		if !ok {
			continue
		}
		if !reflect.DeepEqual(c, o) {
			continue
		}
		changed := false
		for _, m := range c.Volumes {
			if _, changed = updateVolumes[m.Source]; changed {
				break
			}
		}
		for name := range c.Networks.ServiceNetworks {
			if _, changed = updateNetworks[name]; changed {
				break
			}
		}
		if changed {
			continue
		}
		keepServices[name] = struct{}{}
	}
	return keepServices
}

// ServiceSort sort service
func ServiceSort(services map[string]baetyl.ComposeService) []string {
	g := map[string][]string{}
	inDegrees := map[string]int{}
	res := []string{}
	for name, s := range services {
		for _, r := range s.DependsOn {
			if g[r] == nil {
				g[r] = []string{}
			}
			g[r] = append(g[r], name)
		}
		inDegrees[name] = len(s.DependsOn)
	}
	queue := []string{}
	for n, i := range inDegrees {
		if i == 0 {
			queue = append(queue, n)
			inDegrees[n] = -1
		}
	}
	for len(queue) > 0 {
		i := queue[0]
		res = append(res, i)
		queue = queue[1:]
		for _, v := range g[i] {
			inDegrees[v]--
			if inDegrees[v] == 0 {
				inDegrees[v] = -1
				queue = append(queue, v)
			}
		}
	}
	return res
}
