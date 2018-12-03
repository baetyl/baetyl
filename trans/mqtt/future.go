package mqtt

import (
	"time"

	"github.com/256dpi/gomqtt/client/future"
)

const errorKey = "error"

// A Future is a low-level future type that can be extended to transport
// custom information.
type Future struct {
	f *future.Future
}

// NewFuture will return a new Future.
func NewFuture() *Future {
	return &Future{
		f: future.New(),
	}
}

// Wait will wait the given amount of time and return whether the future has been
// completed, canceled or the request timed out.
func (f *Future) Wait(timeout time.Duration) (err error) {
	err = f.f.Wait(timeout)
	if err != nil {
		actaul, ok := f.f.Data.Load(errorKey)
		if ok {
			err = actaul.(error)
		}
	}
	return
}

// Complete will complete the future.
func (f *Future) Complete() {
	f.f.Complete()
}

// Cancel will cancel the future with an error.
func (f *Future) Cancel(err error) {
	if err != nil {
		f.f.Data.Store(errorKey, err)
	}
	f.f.Cancel()
}
