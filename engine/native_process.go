package engine

import (
	"os"
	"syscall"
	"time"

	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/utils"
	"github.com/juju/errors"
)

// NativeSpec spec for native process
type NativeSpec struct {
	Spec
	Exec string
	Argv []string
	Attr os.ProcAttr
}

// NativeProcess native process to run and retry
type NativeProcess struct {
	spec *NativeSpec
	proc *os.Process
	tomb utils.Tomb
}

// NewNativeProcess create a new native process
func NewNativeProcess(s *NativeSpec) *NativeProcess {
	return &NativeProcess{
		spec: s,
	}
}

// Policy returns name
func (w *NativeProcess) Name() string {
	return w.spec.Name
}

// Policy returns restart policy
func (w *NativeProcess) Policy() module.Policy {
	return w.spec.Restart
}

// Start starts process
func (w *NativeProcess) Start(supervising func(Worker) error) error {
	err := w.startProcess()
	if err != nil {
		return errors.Trace(err)
	}
	err = w.tomb.Go(func() error {
		return supervising(w)
	})
	return errors.Trace(err)
}

// Restart starts process
func (w *NativeProcess) Restart() error {
	if !w.tomb.Alive() {
		return errors.Errorf("Process already stopped")
	}
	err := w.startProcess()
	if err != nil {
		return errors.Annotatef(err, "Failed to restart process")
	}
	return nil
}

// Stop stops process
func (w *NativeProcess) Stop() error {
	if !w.tomb.Alive() {
		w.spec.Logger.Debug("Process already stopped")
		return nil
	}
	w.tomb.Kill(nil)
	err := w.stopProcess()
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(w.tomb.Wait())
}

// Wait waits until process is stopped
func (w *NativeProcess) Wait(c chan<- error) {
	defer w.spec.Logger.Info("Process stopped")
	ps, err := w.proc.Wait()
	if err != nil {
		w.spec.Logger.Debug("Failed to wait process:", err)
		c <- err
	}
	c <- errors.Errorf("Process exited: %v", ps)
}

// Dying returns the channel that can be used to wait until process is stopped
func (w *NativeProcess) Dying() <-chan struct{} {
	return w.tomb.Dying()
}

func (w *NativeProcess) startProcess() error {
	proc, err := os.StartProcess(
		w.spec.Exec,
		w.spec.Argv,
		&w.spec.Attr,
	)
	if err != nil {
		return errors.Trace(err)
	}
	w.proc = proc
	w.spec.Logger = w.spec.Logger.WithField("pid", proc.Pid)
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
		w.spec.Logger.Debugln("Process exits by signal:", s, err)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(w.spec.Grace):
		err := w.proc.Kill()
		w.spec.Logger.WithError(err).Warnf("Process killed")
	}
	return nil
}
