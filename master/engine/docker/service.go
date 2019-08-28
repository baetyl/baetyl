package docker

import (
	"fmt"
	"sync"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/master/engine"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	cmap "github.com/orcaman/concurrent-map"
)

const (
	fmtVolumeRW = "%s:%s:rw"
	fmtVolumeRO = "%s:%s:ro"
)

type dockerService struct {
	name      string
	cfg       baetyl.ComposeServiceInfo
	params    containerConfigs
	engine    *dockerEngine
	instances cmap.ConcurrentMap
	log       logger.Logger
}

func (s *dockerService) Name() string {
	return s.name
}

func (s *dockerService) Engine() engine.Engine {
	return s.engine
}

func (s *dockerService) RestartPolicy() baetyl.RestartPolicyInfo {
	return s.cfg.Restart
}

func (s *dockerService) Start() error {
	s.log.Debugf("%s replica: %d", s.name, s.cfg.Replica)
	var instanceName string
	for i := 0; i < s.cfg.Replica; i++ {
		if i == 0 {
			instanceName = s.name
		} else {
			instanceName = fmt.Sprintf("%s.i%d", s.name, i)
		}
		err := s.startInstance(instanceName, nil)
		if err != nil {
			s.Stop()
			return err
		}
	}
	return nil
}

func (s *dockerService) Stop() {
	var wg sync.WaitGroup
	for _, v := range s.instances.Items() {
		wg.Add(1)
		go func(i *dockerInstance, wg *sync.WaitGroup) {
			defer wg.Done()
			i.Close()
		}(v.(*dockerInstance), &wg)
	}
	wg.Wait()
}

func (s *dockerService) Stats() {
	for _, item := range s.instances.Items() {
		instance := item.(*dockerInstance)
		if stats := instance.Stats(); stats["error"] == nil {
			s.engine.SetInstanceStats(s.Name(), instance.Name(), stats, false)
		}
	}
}

func (s *dockerService) StartInstance(instanceName string, dynamicConfig map[string]string) error {
	return s.startInstance(instanceName, dynamicConfig)
}

func (s *dockerService) startInstance(instanceName string, dynamicConfig map[string]string) error {
	s.StopInstance(instanceName)
	params := s.params
	params.config.Env = engine.GenerateInstanceEnv(instanceName, s.params.config.Env, dynamicConfig)
	i, err := s.newInstance(instanceName, params)
	if err != nil {
		return err
	}
	s.instances.Set(instanceName, i)
	return nil
}

func (s *dockerService) StopInstance(instanceName string) error {
	i, ok := s.instances.Get(instanceName)
	if !ok {
		s.log.Debugf("instance (%s) not found", instanceName)
		return nil
	}
	return i.(*dockerInstance).Close()
}
