package engine

import (
	"fmt"
	"time"

	"github.com/baetyl/baetyl/logger"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/jpillora/backoff"
)

// Supervising supervise an instance
func Supervising(instance Instance) error {
	service := instance.Service()
	_engine := service.Engine()
	serviceName := service.Name()
	instanceName := instance.Name()
	defer _engine.DelInstanceStats(serviceName, instanceName, true)
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
		instanceInfo := instance.Info()
		instanceInfo[KeyStatus] = Running
		instanceInfo[KeyStartTime] = time.Now().UTC()
		_engine.SetInstanceStats(serviceName, instanceName, instanceInfo, true)
		go instance.Wait(s)
		select {
		case <-instance.Dying():
			return nil
		case err := <-s:
			switch p.Policy {
			case baetyl.RestartOnFailure:
				// TODO: to test
				if err == nil {
					return nil
				}
				_engine.SetInstanceStats(serviceName, instanceName, NewPartialStatsByStatus(Restarting), true)
				goto RESTART
			case baetyl.RestartAlways:
				_engine.SetInstanceStats(serviceName, instanceName, NewPartialStatsByStatus(Restarting), true)
				goto RESTART
			case baetyl.RestartNo:
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
