package native

import (
	"os"
	"sync"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/sdk-go/openedge"
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

func (s *nativeService) start() error {
	return s.startInstance()
}
