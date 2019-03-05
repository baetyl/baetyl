package main

import (
	"context"
	"fmt"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
	"github.com/jolestar/go-commons-pool"
)

// Function function
type Function struct {
	cfg  FunctionInfo
	ctx  openedge.Context
	pool *pool.ObjectPool
	ids  chan int
	log  logger.Logger
	tomb utils.Tomb
}

// NewFunction creates a new function
func NewFunction(ctx openedge.Context, cfg FunctionInfo) *Function {
	f := &Function{
		cfg: cfg,
		ctx: ctx,
		ids: make(chan int, cfg.Instance.Max),
		log: logger.WithField("function", cfg.Name),
	}
	for index := 1; index <= cfg.Instance.Max; index++ {
		f.ids <- index
	}
	pc := pool.NewDefaultPoolConfig()
	pc.MinIdle = cfg.Instance.Min
	pc.MaxIdle = cfg.Instance.Max
	pc.MaxTotal = cfg.Instance.Max
	pc.MinEvictableIdleTime = cfg.Instance.IdleTime
	pc.TimeBetweenEvictionRuns = cfg.Instance.EvictTime
	f.pool = pool.NewObjectPool(context.Background(), newFactory(f), pc)
	return f
}

// Call calls function to handle message and return result message
func (f *Function) Call(msg *openedge.FunctionMessage) (*openedge.FunctionMessage, error) {
	item, err := f.pool.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	fl := item.(*Instance)
	res, err := fl.Call(msg)
	if err != nil {
		f.log.WithError(err).Errorf("failed to talk with function instance")
		err1 := f.pool.InvalidateObject(context.Background(), item)
		if err1 != nil {
			fl.Close()
			f.log.WithError(err).Errorf("failed to invalidate function instance")
		}
		return nil, err
	}
	f.pool.ReturnObject(context.Background(), item)
	return res, nil
}

// Close close function and stop all function instances
// The function instance will be stopped in three cases:
// 1. function instance returns a system error
// 2. function instance is not invoked for a period of time [TODO]
// 3. function manager is closed
func (f *Function) Close() error {
	f.pool.Close(context.Background())
	f.log.Debugf("function closed")
	f.tomb.Kill(nil)
	return f.tomb.Wait()
}

func (f *Function) getID() (int, error) {
	select {
	case index := <-f.ids:
		return index, nil
	case <-f.tomb.Dying():
		return 0, fmt.Errorf("function closed")
	}
}

func (f *Function) returnID(index int) {
	select {
	case f.ids <- index:
	case <-f.tomb.Dying():
	}
}
