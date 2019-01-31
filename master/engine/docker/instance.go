package docker

import (
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
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

func newDockerInstance(s *dockerService) (*dockerInstance, error) {
	// can only start one instance now, use service name as instance name
	cid, err := s.engine.startContainer(s.info.Name, s.cfgs)
	if err != nil {
		s.log.WithError(err).Warnln("failed to start instance")
		// remove and retry
		s.engine.removeContainerByName(s.info.Name)
		cid, err = s.engine.startContainer(s.info.Name, s.cfgs)
		if err != nil {
			s.log.WithError(err).Warnln("failed to start instance again")
			return nil, err
		}
	}
	i := &dockerInstance{
		id:      cid,
		name:    s.info.Name,
		service: s,
		log:     s.log.WithField("instance", cid),
	}
	return i, i.tomb.Go(func() error {
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

func (i *dockerInstance) Policy() engine.RestartPolicyInfo {
	return i.service.info.Restart
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

func (i *dockerInstance) Dying() <-chan struct{} {
	return i.tomb.Dying()
}

func (i *dockerInstance) Close() error {
	i.tomb.Kill(nil)
	return i.tomb.Wait()
}
