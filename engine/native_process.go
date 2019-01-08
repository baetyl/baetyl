package engine

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/logger"
	"github.com/baidu/openedge/module/master"
	"github.com/baidu/openedge/module/utils"
)

// NativeSpec spec for native process
type NativeSpec struct {
	module  *config.Module
	context *Context
	exec    string
	argv    []string
	attr    os.ProcAttr
}

// NativeProcess native process to run and retry
type NativeProcess struct {
	spec *NativeSpec
	proc *os.Process
	tomb utils.Tomb
	log  *logger.Entry
}

// NewNativeProcess create a new native process
func NewNativeProcess(s *NativeSpec) *NativeProcess {
	return &NativeProcess{
		spec: s,
		log:  logger.WithFields("module", s.module.UniqueName()),
	}
}

// UniqueName unique name of worker
func (w *NativeProcess) UniqueName() string {
	return w.spec.module.UniqueName()
}

// Policy returns restart policy
func (w *NativeProcess) Policy() config.Policy {
	return w.spec.module.Restart
}

// Start starts process
func (w *NativeProcess) Start(supervising func(Worker) error) error {
	err := w.startProcess()
	if err != nil {
		return err
	}
	err = w.tomb.Go(func() error {
		return supervising(w)
	})
	return err
}

// Restart starts process
func (w *NativeProcess) Restart() error {
	if !w.tomb.Alive() {
		return fmt.Errorf("process already stopped")
	}
	err := w.startProcess()
	if err != nil {
		return fmt.Errorf("failed to restart process: %s", err.Error())
	}
	return nil
}

// Stop stops process
func (w *NativeProcess) Stop() error {
	if !w.tomb.Alive() {
		w.log.Debugf("process already stopped")
		return nil
	}
	w.tomb.Kill(nil)
	err := w.stopProcess()
	if err != nil {
		return err
	}
	return w.tomb.Wait()
}

// Wait waits until process is stopped
func (w *NativeProcess) Wait(c chan<- error) {
	defer w.log.Infof("process stopped")
	ps, err := w.proc.Wait()
	if err != nil {
		w.log.Debugln("failed to wait process:", err)
		c <- err
	}
	c <- fmt.Errorf("process exited: %v", ps)
}

// Dying returns the channel that can be used to wait until process is stopped
func (w *NativeProcess) Dying() <-chan struct{} {
	return w.tomb.Dying()
}

// Stats returns the stats of docker container
func (w *NativeProcess) Stats() (*master.ModuleStats, error) {
	// TODO: to implement
	return &master.ModuleStats{}, nil
}

func (w *NativeProcess) startProcess() error {
	proc, err := os.StartProcess(
		w.spec.exec,
		w.spec.argv,
		&w.spec.attr,
	)
	if err != nil {
		return err
	}
	w.proc = proc
	w.log = w.log.WithFields("pid", strconv.Itoa(proc.Pid))
	return nil
}

func (w *NativeProcess) stopProcess() error {
	if w.proc == nil {
		return nil
	}
	w.proc.Signal(syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		s, err := w.proc.Wait()
		w.log.Debugln("process exits by signal:", s, err)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(w.spec.context.Grace):
		err := w.proc.Kill()
		w.log.WithError(err).Warnf("process killed")
	}
	return nil
}
