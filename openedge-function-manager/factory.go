package main

import (
	"context"

	pool "github.com/jolestar/go-commons-pool"
)

// factory factory for function
type factory struct {
	*Function
}

func newFactory(f *Function) *factory {
	return &factory{Function: f}
}

/**
 * Create a pointer to an instance that can be served by the
 * pool and wrap it in a PooledObject to be managed by the pool.
 *
 * return error if there is a problem creating a new instance,
 *    this will be propagated to the code requesting an object.
 */
func (f *factory) MakeObject(_ context.Context) (*pool.PooledObject, error) {
	f.log.Debugf("create function instance")
	i, err := f.NewInstance()
	if err != nil {
		return nil, err
	}
	return pool.NewPooledObject(i), nil
}

/**
 * Destroys an instance no longer needed by the pool.
 */
func (f *factory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	f.log.Debugf("destroy function instance")
	return object.Object.(*Instance).Close()
}

/**
 * Ensures that the instance is safe to be returned by the pool.
 *
 * return false if object is not valid and should
 *         be dropped from the pool, true otherwise.
 */
func (f *factory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	return true
}

/**
 * Reinitialize an instance to be returned by the pool.
 *
 * return error if there is a problem activating object,
 *    this error may be swallowed by the pool.
 */
func (f *factory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}

/**
 * Uninitialize an instance to be returned to the idle object pool.
 *
 * return error if there is a problem passivating obj,
 *    this exception may be swallowed by the pool.
 */
func (f *factory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}
