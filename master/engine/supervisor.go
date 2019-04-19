package engine

import (
	"fmt"
	"time"

	"github.com/baidu/openedge/logger"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/jpillora/backoff"
)

// Supervising supervise an instance
func Supervising(instance Instance) error {
	service := instance.Service()
	_engine := service.Engine()
	serviceName := service.Name()
	instanceName := instance.Name()
	defer _engine.DelInstanceStats(serviceName, instanceName)
	defer instance.Stop()

	c := 0
	p := instance.Service().RestartPolicy()
	b := &backoff.Backoff{
		Min:    p.Backoff.Min,
		Max:    p.Backoff.Max,
		Factor: p.Backoff.Factor,
	}
	l := logger.Global.WithField("service", serviceName).WithField("instance", instanceName)
	s := make(chan error, 1)
	for {
		_engine.AddInstanceStats(serviceName, instanceName, PartialStats{
			KeyID:        instance.ID(),
			KeyName:      instance.Name(),
			KeyStatus:    Running,
			KeyStartTime: time.Now().UTC(),
		})
		go instance.Wait(s)
		select {
		case <-instance.Dying():
			return nil
		case err := <-s:
			switch p.Policy {
			case openedge.RestartOnFailure:
				// TODO: to test
				if err == nil {
					return nil
				}
				_engine.AddInstanceStats(serviceName, instanceName, NewPartialStatsByStatus(Restarting))
				goto RESTART
			case openedge.RestartAlways:
				_engine.AddInstanceStats(serviceName, instanceName, NewPartialStatsByStatus(Restarting))
				goto RESTART
			case openedge.RestartNo:
				// TODO: to test
				return nil
			default:
				l.Errorf("Restart policy (%s) invalid", p.Policy)
				return fmt.Errorf("Restart policy invalid")
			}
		}

	RESTART:
		c++
		if p.Retry.Max > 0 && c > p.Retry.Max {
			l.Errorf("retry too much (%d)", c)
			return fmt.Errorf("retry too much")
		}

		select {
		case <-time.After(b.Duration()):
		case <-instance.Dying():
			return nil
		}

		err := instance.Restart()
		if err != nil {
			l.Errorf("failed to restart module, keep to restart")
			goto RESTART
		}
	}
}
