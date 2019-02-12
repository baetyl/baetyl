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
	process *os.Process
	service *nativeService
	log     logger.Logger
	tomb    utils.Tomb
}

func (s *nativeService) startInstance() error {
	// TODO: support multiple instances
	// can only start one instance now, use service name as instance name
	p, err := s.engine.startProcess(s.info.Name, s.cfgs)
	if err != nil {
		s.log.WithError(err).Warnln("failed to start instance")
		// retry
		p, err = s.engine.startProcess(s.info.Name, s.cfgs)
		if err != nil {
			s.log.WithError(err).Warnln("failed to start instance again")
			return err
		}
	}
	i := &nativeInstance{
		name:    s.info.Name,
		process: p,
		service: s,
		log:     s.log.WithField("instance", p.Pid),
	}
	s.instances.Set(s.info.Name, i)
	return i.tomb.Go(func() error {
		return engine.Supervising(i)
	})
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

func (i *nativeInstance) Policy() engine.RestartPolicyInfo {
	return i.service.info.Restart
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
	p, err := i.service.engine.startProcess(i.service.info.Name, i.service.cfgs)
	if err != nil {
		i.log.WithError(err).Errorf("failed to restart instance")
		return err
	}
	i.process = p
	return nil
}

func (i *nativeInstance) Stop() {
	err := i.service.engine.stopProcess(i.process)
	if err != nil {
		i.log.WithError(err).Errorf("failed to stop instance")
	}
}

func (i *nativeInstance) Dying() <-chan struct{} {
	return i.tomb.Dying()
}

func (i *nativeInstance) Close() error {
	i.tomb.Kill(nil)
	return i.tomb.Wait()
}
