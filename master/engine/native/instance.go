package native

import (
	"os"
	"strconv"
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
)

// Instance instance of service
type nativeInstance struct {
	engine.InstanceStats
	service *nativeService
	params  processConfigs
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
		service: s,
		params:  params,
		log:     log.WithField("pid", p.Pid),
	}
	i.SetStats(map[string]interface{}{
		"id":         strconv.Itoa(p.Pid),
		"name":       name,
		"status":     engine.Running,
		"start_time": time.Now().UTC(),
		"process":    p,
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

func (i *nativeInstance) Log() logger.Logger {
	return i.log
}

func (i *nativeInstance) Policy() openedge.RestartPolicyInfo {
	return i.service.cfg.Restart
}

func (i *nativeInstance) Wait(s chan<- error) {
	defer i.log.Infof("instance stopped")
	err := i.service.engine.waitProcess(i.Stat("process").(*os.Process))
	s <- err
	i.SetStatus(engine.Exited)
}

func (i *nativeInstance) Restart() error {
	i.SetStatus(engine.Restarting)

	p, err := i.service.engine.startProcess(i.params)
	if err != nil {
		i.log.WithError(err).Errorf("failed to restart instance")
		return err
	}
	i.SetStats(map[string]interface{}{
		"id":         strconv.Itoa(p.Pid),
		"status":     engine.Running,
		"start_time": time.Now().UTC(),
		"process":    p,
	})
	i.log = i.log.WithField("pid", p.Pid)
	i.log.Infof("instance restarted")
	return nil
}

func (i *nativeInstance) Stop() {
	i.log.Infof("to stop instance")
	err := i.service.engine.stopProcess(i.Stat("process").(*os.Process))
	if err != nil {
		i.log.Debugf("failed to stop instance: %s", err.Error())
	}
	i.SetStatus(engine.Dead)
	i.service.instances.Remove(i.Name())
}

func (i *nativeInstance) Dying() <-chan struct{} {
	return i.tomb.Dying()
}

func (i *nativeInstance) Close() error {
	i.log.Infof("to close instance")
	i.tomb.Kill(nil)
	return i.tomb.Wait()
}
