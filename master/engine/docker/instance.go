package docker

import (
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
)

// Instance instance of service
type dockerInstance struct {
	engine.InstanceStats
	service *dockerService
	log     logger.Logger
	tomb    utils.Tomb
}

func (s *dockerService) newInstance(name string, params containerConfigs) (*dockerInstance, error) {
	log := s.log.WithField("instance", name)
	cid, err := s.engine.startContainer(name, params)
	if err != nil {
		log.WithError(err).Warnln("failed to start instance, clean and retry")
		// remove and retry
		s.engine.removeContainerByName(name)
		cid, err = s.engine.startContainer(name, params)
		if err != nil {
			log.WithError(err).Warnln("failed to start instance again")
			return nil, err
		}
	}
	i := &dockerInstance{
		service: s,
		log:     log.WithField("cid", cid[:12]),
	}
	i.SetStats(map[string]interface{}{
		"id":   cid,
		"name": name,
	})
	err = i.tomb.Go(func() error {
		return engine.Supervising(i)
	})
	if err != nil {
		i.Close()
		return nil, err
	}
	i.log.Infof("instance started")
	return i, nil
}

func (i *dockerInstance) Log() logger.Logger {
	return i.log
}

func (i *dockerInstance) Policy() openedge.RestartPolicyInfo {
	return i.service.cfg.Restart
}

func (i *dockerInstance) State() openedge.InstanceStatus {
	s, err := i.service.engine.statsContainer(i.ID())
	if err != nil {
		i.log.WithError(err).Errorf("failed to stats instance")
		i.SetStatus(engine.Offline)
	} else {
		i.SetStats(s)
	}
	return i.Stats()
}

func (i *dockerInstance) Wait(s chan<- error) {
	defer i.log.Infof("instance stopped")
	err := i.service.engine.waitContainer(i.ID())
	s <- err
}

func (i *dockerInstance) Restart() error {
	err := i.service.engine.restartContainer(i.ID())
	if err != nil {
		i.log.WithError(err).Errorf("failed to restart instance")
		return err
	}
	i.log.Infof("instance restarted")
	return nil
}

func (i *dockerInstance) Stop() {
	i.log.Infof("to stop instance")
	err := i.service.engine.stopContainer(i.ID())
	if err != nil {
		i.log.WithError(err).Errorf("failed to stop instance")
	}
	i.service.engine.removeContainer(i.ID())
	i.service.instances.Remove(i.Name())
}

func (i *dockerInstance) Dying() <-chan struct{} {
	return i.tomb.Dying()
}

func (i *dockerInstance) Close() error {
	i.log.Infof("to close instance")
	i.tomb.Kill(nil)
	return i.tomb.Wait()
}
