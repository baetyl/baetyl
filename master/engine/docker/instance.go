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

func (s *dockerService) startInstance() error {
	// TODO: support multiple instances
	// can only start one instance now, use service name as instance name
	cid, err := s.engine.startContainer(s.cfg.Name, s.params)
	if err != nil {
		s.log.WithError(err).Warnln("failed to start instance, clean and retry")
		// remove and retry
		s.engine.removeContainerByName(s.cfg.Name)
		cid, err = s.engine.startContainer(s.cfg.Name, s.params)
		if err != nil {
			s.log.WithError(err).Warnln("failed to start instance again")
			return err
		}
	}
	i := &dockerInstance{
		id:      cid,
		name:    s.cfg.Name,
		service: s,
		log:     s.log.WithField("instance", cid[:12]),
	}
	s.instances.Set(s.cfg.Name, i)
	return i.tomb.Go(func() error {
		return engine.Supervising(i)
	})
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
		i.log.WithError(err).Errorf("failed to restart instance, to retry")
	}
	return err
}

func (i *dockerInstance) Stop() {
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
	i.tomb.Kill(nil)
	return i.tomb.Wait()
}
