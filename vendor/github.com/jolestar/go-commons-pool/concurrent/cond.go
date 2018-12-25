package concurrent

import (
	"context"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
)

// TimeoutCond is a sync.Cond  improve for support wait timeout.
type TimeoutCond struct {
	L          sync.Locker
	signal     chan int
	hasWaiters uint64
}

// NewTimeoutCond return a new TimeoutCond
func NewTimeoutCond(l sync.Locker) *TimeoutCond {
	cond := TimeoutCond{L: l, signal: make(chan int, 0)}
	return &cond
}

func (cond *TimeoutCond) addWaiter() {
	v := atomic.AddUint64(&cond.hasWaiters, 1)
	if v == 0 {
		panic("too many waiters; max is " + strconv.FormatUint(math.MaxUint64, 10))
	}
}

func (cond *TimeoutCond) removeWaiter() {
	// Decrement. See notes here: https://godoc.org/sync/atomic#AddUint64
	v := atomic.AddUint64(&cond.hasWaiters, ^uint64(0))

	if v == math.MaxUint64 {
		panic("removeWaiter called more than once after addWaiter")
	}
}

// HasWaiters queries whether any goroutine are waiting on this condition
func (cond *TimeoutCond) HasWaiters() bool {
	return atomic.LoadUint64(&cond.hasWaiters) > 0
}

// Wait waits for a signal, or for the context do be done. Returns true if signaled.
func (cond *TimeoutCond) Wait(ctx context.Context) bool {
	cond.addWaiter()
	//copy signal in lock, avoid data race with Interrupt
	ch := cond.signal
	//wait should unlock mutex,  if not will cause deadlock
	cond.L.Unlock()
	defer cond.removeWaiter()
	defer cond.L.Lock()

	select {
	case _, ok := <-ch:
		return !ok
	case <-ctx.Done():
		return false
	}
}

// Signal wakes one goroutine waiting on c, if there is any.
func (cond *TimeoutCond) Signal() {
	select {
	case cond.signal <- 1:
	default:
	}
}

// Interrupt goroutine wait on this TimeoutCond
func (cond *TimeoutCond) Interrupt() {
	cond.L.Lock()
	defer cond.L.Unlock()
	close(cond.signal)
	cond.signal = make(chan int, 0)
}
