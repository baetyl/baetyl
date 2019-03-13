package native

import (
	"os"
	"strconv"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
)

// Instance instance of service
type nativeInstance struct {
	name    string
	process *os.Process
	service *nativeService
	params  processConfigs
	log     logger.Logger
	tomb    utils.Tomb
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
		process: p,
		service: s,
		params:  params,
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
	return strconv.Itoa(i.process.Pid)
}

func (i *nativeInstance) Name() string {
	return i.name
}

func (i *nativeInstance) Log() logger.Logger {
	return i.log
}

func (i *nativeInstance) Policy() openedge.RestartPolicyInfo {
	return i.service.cfg.Restart
}

func (i *nativeInstance) Wait(s chan<- error) {
	defer i.log.Infof("instance stopped")
	err := i.service.engine.waitProcess(i.process)
	s <- err
}

func (i *nativeInstance) Restart() error {
	// err := i.service.engine.stopProcess(i.process)
	// if err != nil {
	// 	i.log.WithError(err).Errorf("failed to stop instance")
	// }
	p, err := i.service.engine.startProcess(i.params)
	if err != nil {
		i.log.WithError(err).Errorf("failed to restart instance")
		return err
	}
	i.process = p
	i.log = i.log.WithField("pid", p.Pid)
	i.log.Infof("instance restarted")
	return nil
}

func (i *nativeInstance) Stop() {
	i.log.Infof("to stop instance")
	err := i.service.engine.stopProcess(i.process)
	if err != nil {
		i.log.WithError(err).Errorf("failed to stop instance")
		return
	}

}

func (i *nativeInstance) Dying() <-chan struct{} {
	return i.tomb.Dying()
}

func (i *nativeInstance) Close() error {
	i.log.Infof("to close instance")
	i.tomb.Kill(nil)
	return i.tomb.Wait()
}
