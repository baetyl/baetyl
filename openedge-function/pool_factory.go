package main

import (
	"context"

	pool "github.com/jolestar/go-commons-pool"
)

// functionFactory factory for function
type functionFactory struct {
	*Function
}

func newFuncionFactory(f *Function) *functionFactory {
	return &functionFactory{Function: f}
}

/**
 * Create a pointer to an instance that can be served by the
 * pool and wrap it in a PooledObject to be managed by the pool.
 *
 * return error if there is a problem creating a new instance,
 *    this will be propagated to the code requesting an object.
 */
func (f *functionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	fl, err := f.newFunclet()
	if err != nil {
		return nil, err
	}
	return pool.NewPooledObject(fl), nil
}

/**
 * Destroys an instance no longer needed by the pool.
 */
func (f *functionFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	object.Object.(*funclet).Close()
	return nil
}

/**
 * Ensures that the instance is safe to be returned by the pool.
 *
 * return false if object is not valid and should
 *         be dropped from the pool, true otherwise.
 */
func (f *functionFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	return true
}

/**
 * Reinitialize an instance to be returned by the pool.
 *
 * return error if there is a problem activating object,
 *    this error may be swallowed by the pool.
 */
func (f *functionFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}

/**
 * Uninitialize an instance to be returned to the idle object pool.
 *
 * return error if there is a problem passivating obj,
 *    this exception may be swallowed by the pool.
 */
func (f *functionFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}
