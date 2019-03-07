package docker

import (
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
)

// Instance instance of service
type dockerInstance struct {
	id      string
	name    string
	service *dockerService
	log     logger.Logger
	tomb    utils.Tomb
}

func (s *dockerService) newInstance(name string, params containerConfigs) (*dockerInstance, error) {
	log := s.log.WithField("instance", name)
	containerName := s.cfg.Name
	if name != "" {
		containerName = containerName + "." + name
	}
	cid, err := s.engine.startContainer(containerName, params)
	if err != nil {
		log.WithError(err).Warnln("failed to start instance, clean and retry")
		// remove and retry
		s.engine.removeContainerByName(containerName)
		cid, err = s.engine.startContainer(containerName, params)
		if err != nil {
			log.WithError(err).Warnln("failed to start instance again")
			return nil, err
		}
	}
	i := &dockerInstance{
		id:      cid,
		name:    name,
		service: s,
		log:     log.WithField("cid", cid[:12]),
	}
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

func (i *dockerInstance) ID() string {
	return i.id
}

func (i *dockerInstance) Name() string {
	return i.name
}

func (i *dockerInstance) Log() logger.Logger {
	return i.log
}

func (i *dockerInstance) Policy() openedge.RestartPolicyInfo {
	return i.service.cfg.Restart
}

func (i *dockerInstance) Wait(s chan<- error) {
	defer i.log.Infof("instance stopped")
	err := i.service.engine.waitContainer(i.id)
	s <- err
}

func (i *dockerInstance) Restart() error {
	err := i.service.engine.restartContainer(i.id)
	if err != nil {
		i.log.WithError(err).Errorf("failed to restart instance")
		return err
	}
	i.log.Infof("instance restarted")
	return nil
}

func (i *dockerInstance) Stop() {
	i.log.Infof("to stop instance")
	err := i.service.engine.stopContainer(i.id)
	if err != nil {
		i.log.WithError(err).Errorf("failed to stop instance")
	}
	i.service.engine.removeContainer(i.id)
}

func (i *dockerInstance) Dying() <-chan struct{} {
	return i.tomb.Dying()
}

func (i *dockerInstance) Close() error {
	i.log.Infof("to close instance")
	i.tomb.Kill(nil)
	return i.tomb.Wait()
}
