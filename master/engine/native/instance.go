package native

import (
	"os"
	"strconv"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/utils"
)

// Instance instance of service
type nativeInstance struct {
	name    string
	service *nativeService
	params  processConfigs
	proc    *os.Process
	tomb    utils.Tomb
	log     logger.Logger
}

func (s *nativeService) newInstance(name string, params processConfigs) (*nativeInstance, error) {
	log := s.log.WithField("instance", name)
	p, err := s.engine.startProcess(params)
	if err != nil {
		log.WithError(err).Warnf("failed to start instance")
		// retry
		p, err = s.engine.startProcess(params)
		if err != nil {
			log.WithError(err).Warnf("failed to start instance again")
			return nil, err
		}
	}
	i := &nativeInstance{
		name:    name,
		service: s,
		params:  params,
		proc:    p,
		log:     log.WithField("pid", p.Pid),
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

func (i *nativeInstance) ID() string {
	return strconv.Itoa(i.proc.Pid)
}

func (i *nativeInstance) Name() string {
	return i.name
}

func (i *nativeInstance) Service() engine.Service {
	return i.service
}

func (i *nativeInstance) Wait(s chan<- error) {
	defer i.log.Infof("instance stopped")
	err := i.service.engine.waitProcess(i.proc)
	s <- err
}

func (i *nativeInstance) Restart() error {
	p, err := i.service.engine.startProcess(i.params)
	if err != nil {
		i.log.WithError(err).Errorf("failed to restart instance")
		return err
	}
	i.proc = p
	i.log = i.log.WithField("pid", p.Pid)
	i.log.Infof("instance restarted")
	return nil
}

func (i *nativeInstance) Stop() {
	i.log.Infof("to stop instance")
	err := i.service.engine.stopProcess(i.proc)
	if err != nil {
		i.log.Debugf("failed to stop instance: %s", err.Error())
	}
	i.service.instances.Remove(i.name)
}

func (i *nativeInstance) Dying() <-chan struct{} {
	return i.tomb.Dying()
}

func (i *nativeInstance) Close() error {
	i.log.Infof("to close instance")
	i.tomb.Kill(nil)
	return i.tomb.Wait()
}
