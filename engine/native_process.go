package engine

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/utils"
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

// Name returns name
func (w *NativeProcess) Name() string {
	return w.spec.Name
}

// Policy returns restart policy
func (w *NativeProcess) Policy() config.Policy {
	return w.spec.Restart
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
		w.spec.Logger.Debugf("process already stopped")
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
	defer w.spec.Logger.Infof("process stopped")
	ps, err := w.proc.Wait()
	if err != nil {
		w.spec.Logger.Debugln("failed to wait process:", err)
		c <- err
	}
	c <- fmt.Errorf("process exited: %v", ps)
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
		return err
	}
	w.proc = proc
	w.spec.Logger = w.spec.Logger.WithFields("pid", strconv.Itoa(proc.Pid))
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
		w.spec.Logger.Debugln("process exits by signal:", s, err)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(w.spec.Grace):
		err := w.proc.Kill()
		w.spec.Logger.WithError(err).Warnf("process killed")
	}
	return nil
}
