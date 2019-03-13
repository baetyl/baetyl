package native

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/baidu/openedge/logger"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/orcaman/concurrent-map"
)

const packageConfigPath = "package.yml"

type packageConfig struct {
	Entry string `yaml:"entry" json:"entry"`
}

type nativeService struct {
	cfg       openedge.ServiceInfo
	params    processConfigs
	engine    *nativeEngine
	instances cmap.ConcurrentMap
	wdir      string
	log       logger.Logger
}

func (s *nativeService) Name() string {
	return s.cfg.Name
}

func (s *nativeService) Stats() openedge.ServiceStatus {
	return openedge.ServiceStatus{}
}

func (s *nativeService) Start() error {
	s.log.Debugf("%s replica: %d", s.cfg.Name, s.cfg.Replica)
	var instanceName string
	for i := 0; i < s.cfg.Replica; i++ {
		if i == 0 {
			instanceName = ""
		} else {
			instanceName = strconv.Itoa(i)
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

func (s *nativeService) StartInstance(instanceName string, dynamicConfig map[string]string) error {
	return s.startInstance(instanceName, dynamicConfig)
}

func (s *nativeService) startInstance(instanceName string, dynamicConfig map[string]string) error {
	s.StopInstance(instanceName)
	params := s.params
	if dynamicConfig != nil {
		// now only support to use env to pass dynamic config
		params.env = []string{}
		for k, v := range dynamicConfig {
			params.env = append(params.env, fmt.Sprintf("%s=%s", k, v))
		}
	}
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
	s.instances.Remove(instanceName)
	return i.(*nativeInstance).Close()
}
