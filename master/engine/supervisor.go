package engine

import (
	"fmt"
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/jpillora/backoff"
)

// Supervisee supervisee supervised by supervisor
type Supervisee interface {
	Log() logger.Logger
	Policy() RestartPolicyInfo
	Wait(w chan<- error)
	Dying() <-chan struct{}
	Restart() error
	Stop()
}

// Supervising supervise a supervisee
func Supervising(supervisee Supervisee) error {
	defer supervisee.Stop()

	c := 0
	p := supervisee.Policy()
	b := &backoff.Backoff{
		Min:    p.Backoff.Min,
		Max:    p.Backoff.Max,
		Factor: p.Backoff.Factor,
	}
	l := supervisee.Log()
	s := make(chan error, 1)
	for {
		go supervisee.Wait(s)
		select {
		case <-supervisee.Dying():
			return nil
		case err := <-s:
			switch p.Policy {
			case RestartOnFailure:
				if err == nil {
					return nil
				}
				goto RESTART
			case RestartAlways:
				goto RESTART
			case RestartNo:
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
		case <-supervisee.Dying():
			return nil
		}

		err := supervisee.Restart()
		if err != nil {
			l.Errorf("failed to restart module, keep to restart")
			goto RESTART
		}
	}
}
