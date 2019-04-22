package docker

import (
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/utils"
)

type attribute struct {
	Name      string `yaml:"name" json:"name"`
	Container struct {
		ID   string `yaml:"id" json:"id"`
		Name string `yaml:"name" json:"name"`
	} `yaml:"container" json:"container"`
}

func (a attribute) toPartialStats() engine.PartialStats {
	return engine.PartialStats{
		engine.KeyName: a.Name,
		"container":    a.Container,
	}
}

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
		id:      cid,
		name:    name,
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

func (i *dockerInstance) Service() engine.Service {
	return i.service
}

func (i *dockerInstance) Name() string {
	return i.name
}

func (i *dockerInstance) Info() engine.PartialStats {
	var attr attribute
	attr.Name = i.name
	attr.Container.ID = i.id
	attr.Container.Name = i.name
	return attr.toPartialStats()
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
	i.service.instances.Remove(i.name)
}

func (i *dockerInstance) Dying() <-chan struct{} {
	return i.tomb.Dying()
}

func (i *dockerInstance) Close() error {
	i.log.Infof("to close instance")
	i.tomb.Kill(nil)
	return i.tomb.Wait()
}
