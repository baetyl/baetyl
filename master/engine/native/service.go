package native

import (
	"fmt"
	"os"
	"sync"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/master/engine"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	cmap "github.com/orcaman/concurrent-map"
)

const packageConfigPath = "package.yml"

type packageConfig struct {
	Entry string `yaml:"entry" json:"entry"`
}

type nativeService struct {
	name      string
	cfg       baetyl.ComposeServiceInfo
	params    processConfigs
	engine    *nativeEngine
	instances cmap.ConcurrentMap
	wdir      string
	log       logger.Logger
}

func (s *nativeService) Name() string {
	return s.name
}

func (s *nativeService) Engine() engine.Engine {
	return s.engine
}

func (s *nativeService) RestartPolicy() baetyl.RestartPolicyInfo {
	return s.cfg.Restart
}

func (s *nativeService) Start() error {
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

func (s *nativeService) Stop() {
	defer os.RemoveAll(s.params.pwd)
	var wg sync.WaitGroup
	for _, v := range s.instances.Items() {
		wg.Add(1)
		go func(i *nativeInstance, wg *sync.WaitGroup) {
			defer wg.Done()
			i.Close()
		}(v.(*nativeInstance), &wg)
	}
	wg.Wait()
}

func (s *nativeService) Stats() {
	for _, item := range s.instances.Items() {
		instance := item.(*nativeInstance)
		if stats := instance.Stats(); stats["error"] == nil {
			s.engine.SetInstanceStats(s.Name(), instance.Name(), stats, false)
		}
	}
}

func (s *nativeService) StartInstance(instanceName string, dynamicConfig map[string]string) error {
	return s.startInstance(instanceName, dynamicConfig)
}

func (s *nativeService) startInstance(instanceName string, dynamicConfig map[string]string) error {
	s.StopInstance(instanceName)
	params := s.params
	params.argv = append([]string{instanceName}, s.params.argv...)
	params.env = engine.GenerateInstanceEnv(instanceName, s.params.env, dynamicConfig)
	i, err := s.newInstance(instanceName, params)
	if err != nil {
		return err
	}
	s.instances.Set(instanceName, i)
	return nil
}

func (s *nativeService) StopInstance(instanceName string) error {
	i, ok := s.instances.Get(instanceName)
	if !ok {
		s.log.Debugf("instance (%s) not found", instanceName)
		return nil
	}
	return i.(*nativeInstance).Close()
}
