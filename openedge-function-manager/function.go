package main

import (
	"context"
	"fmt"

	"github.com/baidu/openedge/logger"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	pool "github.com/jolestar/go-commons-pool"
)

// Function function
type Function struct {
	p    Producer
	cfg  FunctionInfo
	ids  chan uint32
	pool *pool.ObjectPool
	log  logger.Logger
	tomb utils.Tomb
}

// NewFunction creates a new function
func NewFunction(cfg FunctionInfo, p Producer) *Function {
	f := &Function{
		p:   p,
		cfg: cfg,
		ids: make(chan uint32, cfg.Instance.Max),
		log: logger.WithField("function", cfg.Name),
	}
	for index := 1; index <= cfg.Instance.Max; index++ {
		f.ids <- uint32(index)
	}
	pc := pool.NewDefaultPoolConfig()
	pc.MinIdle = cfg.Instance.Min
	pc.MaxIdle = cfg.Instance.Max
	pc.MaxTotal = cfg.Instance.Max
	pc.MinEvictableIdleTime = cfg.Instance.IdleTime
	pc.TimeBetweenEvictionRuns = cfg.Instance.EvictTime
	f.pool = pool.NewObjectPool(context.Background(), f, pc)
	return f
}

// Call calls function to handle message and return result message
func (f *Function) Call(msg *openedge.FunctionMessage) (*openedge.FunctionMessage, error) {
	item, err := f.pool.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	return f.call(item.(Instance), msg, nil)
}

// CallAsync calls function to handle message and return result message
func (f *Function) CallAsync(msg *openedge.FunctionMessage, cb func(in, out *openedge.FunctionMessage, err error)) error {
	item, err := f.pool.BorrowObject(context.Background())
	if err != nil {
		return err
	}
	go f.call(item.(Instance), msg, cb)
	return nil
}

func (f *Function) call(i Instance, in *openedge.FunctionMessage, c func(in, out *openedge.FunctionMessage, err error)) (*openedge.FunctionMessage, error) {
	out, err := i.Call(in)
	if err != nil {
		f.log.Errorf("failed to talk with function instance: %s", err.Error())
		err1 := f.pool.InvalidateObject(context.Background(), i)
		if err1 != nil {
			i.Close()
			f.log.Errorf("failed to invalidate function instance: %s", err1.Error())
		}
	} else {
		err1 := f.pool.ReturnObject(context.Background(), i)
		if err1 != nil {
			i.Close()
			f.log.Errorf("failed to return function instance: %s", err1.Error())
		}
	}
	if c != nil {
		c(in, out, err)
	}
	return out, err
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

// MakeObject creates a new instance
func (f *Function) MakeObject(_ context.Context) (*pool.PooledObject, error) {
	select {
	case id := <-f.ids:
		i, err := f.p.StartInstance(id)
		if err != nil {
			return nil, err
		}
		return pool.NewPooledObject(i), nil
	case <-f.tomb.Dying():
		return nil, fmt.Errorf("function closed")
	}

}

// DestroyObject close an instance
func (f *Function) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	i := object.Object.(Instance)
	i.Close()
	select {
	case f.ids <- i.ID():
	case <-f.tomb.Dying():
	}
	return nil
}

// ValidateObject not implement
func (f *Function) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	return true
}

// ActivateObject not implement
func (f *Function) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}

// PassivateObject not implement
func (f *Function) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}
