package baetyl

import (
	"context"
	"sync"

	"gopkg.in/tomb.v2"
)

type task struct {
	c context.Context
	m sync.Mutex
	t *tomb.Tomb
}

func (t *task) run(code func(ctx context.Context) error) {
	t.m.Lock()
	if t.t != nil {
		t.t.Kill(nil)
		t.t.Wait()
	}
	var ictx context.Context
	t.t, ictx = tomb.WithContext(t.c)
	t.t.Go(func() error {
		return code(ictx)
	})
	t.m.Unlock()
}
