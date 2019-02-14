package docker

import (
	"sync"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/orcaman/concurrent-map"
)

const (
	fmtVolume   = "%s:/%s"
	fmtVolumeRO = "%s:/%s:ro"
)

type dockerService struct {
	info      engine.ServiceInfo
	cfgs      containerConfigs
	engine    *dockerEngine
	instances cmap.ConcurrentMap
	log       logger.Logger
}

func (s *dockerService) Name() string {
	return s.info.Name
}

func (s *dockerService) Stats() openedge.ServiceStatus {
	instances := s.instances.Items()
	results := make(chan openedge.InstanceStatus, len(instances))

	var wg sync.WaitGroup
	for _, v := range instances {
		wg.Add(1)
		go func(i *dockerInstance, wg *sync.WaitGroup) {
			defer wg.Done()
			status, err := i.service.engine.statsContainer(i.id)
			if err != nil {
				status = openedge.InstanceStatus{"error": err.Error()}
			}
			status["id"] = i.ID()
			status["name"] = i.ID()
			results <- status
		}(v.(*dockerInstance), &wg)
	}
	wg.Wait()
	close(results)
	r := openedge.NewServiceStatus(s.info.Name)
	for i := range results {
		r.Instances = append(r.Instances, i)
	}
	return r
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

func (s *dockerService) start() error {
	return s.startInstance()
}
