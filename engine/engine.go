package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/baidu/openedge/config"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/module"
	"github.com/docker/distribution/uuid"
	"github.com/jpillora/backoff"
	"github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
)

const (
	// ModeNative runs modules in native mode
	ModeNative = "native"
	// ModeDocker runs modules in docker mode
	ModeDocker = "docker"
)

// Context represents the context of engine to execute
type Context struct {
	Mode  string
	Grace time.Duration
}

// Spec common spec
type Spec struct {
	Name    string
	Restart module.Policy
	Logger  *logrus.Entry
	Grace   time.Duration
}

// Inner prepares and creates
type Inner interface {
	Prepare(string) error
	Create(config.Module) (Worker, error)
}

// Worker worker
type Worker interface {
	Name() string
	Policy() module.Policy
	Start(supervising func(Worker) error) error
	Restart() error
	Stop() error
	Wait(w chan<- error)
	Dying() <-chan struct{}
}

// Engine manages all modules
type Engine struct {
	Inner
	auth      map[string]string
	order     []string           // resident module start order
	resident  cmap.ConcurrentMap // resident modules from app.yml
	temporary cmap.ConcurrentMap // temporary modules from function module
	entries   cmap.ConcurrentMap
	log       *logrus.Entry
}

// New creates a new engine
func New(ctx *Context) (*Engine, error) {
	e := &Engine{
		auth:      map[string]string{},
		order:     []string{},
		resident:  cmap.New(),
		temporary: cmap.New(),
		entries:   cmap.New(),
		log:       logger.WithFields("mode", ctx.Mode),
	}
	var err error
	switch ctx.Mode {
	case ModeDocker:
		e.Inner, err = NewDockerEngine(ctx)
	case ModeNative:
		e.Inner, err = NewNativeEngine(ctx)
	default:
		err = fmt.Errorf("mode (%s) not supported", ctx.Mode)
	}
	return e, err
}

// StartAll starts all resident modules
func (e *Engine) StartAll(ms []config.Module) error {
	entries := map[string]struct{}{}
	for _, m := range ms {
		entries[m.Entry] = struct{}{}
	}
	err := e.prepare(entries)
	if err != nil {
		e.log.WithError(err).Warnf("failed to prepare entries")
	}
	for _, m := range ms {
		if _, ok := e.auth[m.Name]; !ok {
			e.auth[m.Name] = uuid.Generate().String()
		}
		m.Env[module.EnvOpenEdgeModuleToken] = e.auth[m.Name]
		worker, err := e.Create(m)
		if err != nil {
			return err
		}
		e.resident.Set(m.Name, worker)
		e.order = append(e.order, m.Name)
		err = worker.Start(e.supervising)
		if err != nil {
			e.log.WithError(err).Errorf("failed to start resident module (%s)", m.Name)
			return err
		}
		e.log.Infof("resident module (%s) started", m.Name)
	}
	e.log.Info("all resident modules started")
	return nil
}

// Authenticate authenticate account
func (e *Engine) Authenticate(username, password string) bool {
	if username == "" || password == "" {
		return false
	}
	p, ok := e.auth[username]
	return ok && p == password
}

// Start starts a temporary module
func (e *Engine) Start(m config.Module) error {
	e.log.Debugln("starting temporary module:", m)
	if !e.entries.Has(m.Entry) {
		err := e.prepare(map[string]struct{}{m.Entry: struct{}{}})
		if err != nil {
			e.log.WithError(err).Warnf("failed to prepare entry of temporary module (%s)", m.Name)
		}
	}
	worker, err := e.Create(m)
	if err != nil {
		e.log.WithError(err).Errorf("failed to create temporary module (%s)", m.Name)
		return err
	}
	old, ok := e.temporary.Get(m.Name)
	if ok {
		e.temporary.Remove(m.Name)
		old.(Worker).Stop()
	}
	err = worker.Start(e.supervising)
	if err != nil {
		worker.Stop()
		e.log.WithError(err).Errorf("failed to start temporary module (%s)", m.Name)
	} else {
		e.temporary.Set(m.Name, worker)
		e.log.Infof("temporary module (%s) started", m.Name)
	}
	return err
}

// Restart restart a temporary module
func (e *Engine) Restart(name string) error {
	old, ok := e.temporary.Get(name)
	if !ok {
		return fmt.Errorf("temporary module (%s) not found", name)
	}
	err := old.(Worker).Restart()
	if err != nil {
		e.log.WithError(err).Errorf("failed to restart temporary module (%s)", name)
	} else {
		e.log.Infof("temporary module (%s) restarted", name)
	}
	return err
}

// Stop stops a temporary module
func (e *Engine) Stop(name string) error {
	old, ok := e.temporary.Get(name)
	if !ok {
		return nil
	}
	defer e.log.Infof("temporary module (%s) stopped", name)
	e.temporary.Remove(name)
	go old.(Worker).Stop()
	return nil
}

// StopAll stops all modules
func (e *Engine) StopAll() {
	for index := len(e.order) - 1; index >= 0; index-- {
		name := e.order[index]
		w, ok := e.resident.Get(name)
		if ok {
			e.resident.Remove(name)
			err := w.(Worker).Stop()
			if err != nil {
				e.log.WithError(err).Errorf("failed to stop resident module (%s)", name)
			} else {
				e.log.Infof("resident module (%s) stopped", name)
			}
		}
	}
	e.order = []string{}
	e.log.Info("all resident modules stopped")
	var wg sync.WaitGroup
	for item := range e.temporary.IterBuffered() {
		e.temporary.Remove(item.Key)
		wg.Add(1)
		go func(w Worker) {
			err := w.Stop()
			if err != nil {
				e.log.WithError(err).Errorf("failed to stop temporary module")
			} else {
				e.log.Infof("temporary module stopped")
			}
			wg.Done()
		}(item.Val.(Worker))
	}
	wg.Wait()
	e.log.Info("all temporary modules stopped")
}

// Prepare prepares entries
func (e *Engine) prepare(entries map[string]struct{}) error {
	type prepared struct {
		name string
		err  error
	}
	results := make(chan prepared, len(entries))
	for key := range entries {
		go func(entry string) {
			results <- prepared{name: entry, err: e.Prepare(entry)}
		}(key)
	}
	message := ""
	for index := 0; index < len(entries); index++ {
		p := <-results
		if p.err == nil {
			e.entries.Set(p.name, struct{}{})
		} else {
			message = message + p.err.Error() + ";"
		}
	}
	if message != "" {
		return fmt.Errorf(message)
	}
	e.log.Infof("entry (%v) prepared", entries)
	return nil
}

func (e *Engine) supervising(w Worker) error {
	defer e.Stop(w.Name())
	r := w.Policy()
	c := make(chan error, 1)
	backoff := &backoff.Backoff{
		Min:    r.Backoff.Min,
		Max:    r.Backoff.Max,
		Factor: r.Backoff.Factor,
	}
	count := 0
	for {
		go w.Wait(c)
		select {
		case <-w.Dying():
			return nil
		case err := <-c:
			switch r.Policy {
			case module.RestartUnlessStopped:
				if err != nil {
					return nil
				}
				goto RESTART
			case module.RestartOnFailure:
				if err == nil {
					return nil
				}
				goto RESTART
			case module.RestartAlways:
				goto RESTART
			case module.RestartNo:
				return nil
			default:
				logger.Errorf("Restart policy (%s) invalid", r.Policy)
				return fmt.Errorf("Restart policy invalid")
			}
		}

	RESTART:
		count++
		if r.Retry.Max > 0 && count > r.Retry.Max {
			logger.Errorf("retry too much (%d)", count)
			return fmt.Errorf("retry too much")
		}

		select {
		case <-time.After(backoff.Duration()):
		case <-w.Dying():
			return nil
		}

		err := w.Restart()
		if err != nil {
			logger.Errorf("failed to restart module, keep to restart")
			goto RESTART
		}
	}
}
