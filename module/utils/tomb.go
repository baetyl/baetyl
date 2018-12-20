package utils

import (
	"fmt"
	"sync"

	tb "gopkg.in/tomb.v2"
)

const (
	ini = int32(0)
	gos = int32(1)
)

// Tomb wraps tomb.Tomb
type Tomb struct {
	t tb.Tomb
	s int32
	m sync.Mutex
}

// Go runs functions in new goroutines.
func (t *Tomb) Go(fs ...func() error) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()
	t.m.Lock()
	defer t.m.Unlock()
	t.s = gos
	for _, f := range fs {
		t.t.Go(f)
	}
	return
}

// Kill puts the tomb in a dying state for the given reason.
func (t *Tomb) Kill(reason error) {
	t.t.Kill(reason)
}

// Dying returns the channel that can be used to wait until
// t.Kill is called.
func (t *Tomb) Dying() <-chan struct{} {
	return t.t.Dying()
}

// Wait blocks until all goroutines have finished running, and
// then returns the reason for their death.
//
// If tomb does not start any goroutine, return quickly
func (t *Tomb) Wait() (err error) {
	t.m.Lock()
	if t.s == gos {
		err = t.t.Wait()
	}
	t.m.Unlock()
	return
}

// Alive returns true if the tomb is not in a dying or dead state.
func (t *Tomb) Alive() bool {
	return t.t.Alive()
}
