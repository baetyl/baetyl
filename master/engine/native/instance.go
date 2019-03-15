package native

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
)

// State state of instance
type State string

// all states
const (
	Created    State = "created"    // 已创建
	Running    State = "running"    // 运行中
	Paused     State = "paused"     // 已暂停
	Restarting State = "restarting" // 重启中
	Removing   State = "removing"   // 退出中
	Exited     State = "exited"     // 已退出
	Dead       State = "dead"       // 未启动（默认值）
	Offline    State = "offline"    // 离线（同核心的状态）
)

// Instance instance of service
type nativeInstance struct {
	name    string
	service *nativeService
	params  processConfigs
	tomb    utils.Tomb
	log     logger.Logger

	mutex     sync.RWMutex
	state     State
	process   *os.Process
	startTime time.Time
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
		name:      name,
		process:   p,
		service:   s,
		params:    params,
		state:     Running,
		startTime: time.Now().UTC(),
		log:       log.WithField("pid", p.Pid),
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

func (i *nativeInstance) Log() logger.Logger {
	return i.log
}

func (i *nativeInstance) Policy() openedge.RestartPolicyInfo {
	return i.service.cfg.Restart
}

func (i *nativeInstance) State() openedge.InstanceStatus {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	return openedge.InstanceStatus{
		"status":     i.state,
		"id":         strconv.Itoa(i.process.Pid),
		"name":       i.name,
		"start_time": i.startTime,
	}
}

func (i *nativeInstance) setState(s State) {
	i.mutex.Lock()
	i.state = s
	i.mutex.Unlock()
}

func (i *nativeInstance) getProcess() *os.Process {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.process
}

func (i *nativeInstance) setProcess(p *os.Process) {
	i.mutex.Lock()
	i.process = p
	i.startTime = time.Now().UTC()
	i.mutex.Unlock()
}

func (i *nativeInstance) Wait(s chan<- error) {
	defer i.log.Infof("instance stopped")
	err := i.service.engine.waitProcess(i.getProcess())
	s <- err
	i.setState(Exited)
}

func (i *nativeInstance) Restart() error {
	i.setState(Restarting)

	p, err := i.service.engine.startProcess(i.params)
	if err != nil {
		i.log.WithError(err).Errorf("failed to restart instance")
		return err
	}

	i.setProcess(p)
	i.setState(Running)
	i.log = i.log.WithField("pid", p.Pid)
	i.log.Infof("instance restarted")
	return nil
}

func (i *nativeInstance) Stop() {
	i.log.Infof("to stop instance")
	err := i.service.engine.stopProcess(i.getProcess())
	if err != nil {
		i.log.Debugf("failed to stop instance: %s", err.Error())
	}
	i.setState(Dead)
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
