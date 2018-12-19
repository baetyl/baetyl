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
	tb.Tomb
	s int32
	m sync.Mutex
}

// Gos goes functions
func (t *Tomb) Gos(fs ...func() error) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()
	t.m.Lock()
	defer t.m.Unlock()
	t.s = gos
	for _, f := range fs {
		t.Go(f)
	}
	return
}

// Kill signal
func (t *Tomb) Kill() {
	t.Tomb.Kill(nil)
}

// KillWith signal with error
func (t *Tomb) KillWith(err error) {
	t.Tomb.Kill(err)
}

// Wait waits goroutines terminated
func (t *Tomb) Wait() (err error) {
	t.m.Lock()
	if t.s == gos {
		err = t.Tomb.Wait()
	}
	t.m.Unlock()
	return
}
