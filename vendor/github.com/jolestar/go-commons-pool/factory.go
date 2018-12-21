package pool

import (
	"context"
	"errors"
)

// PooledObjectFactory is factory interface for ObjectPool
type PooledObjectFactory interface {

	/**
	 * Create a pointer to an instance that can be served by the
	 * pool and wrap it in a PooledObject to be managed by the pool.
	 *
	 * return error if there is a problem creating a new instance,
	 *    this will be propagated to the code requesting an object.
	 */
	MakeObject(ctx context.Context) (*PooledObject, error)

	/**
	 * Destroys an instance no longer needed by the pool.
	 */
	DestroyObject(ctx context.Context, object *PooledObject) error

	/**
	 * Ensures that the instance is safe to be returned by the pool.
	 *
	 * return false if object is not valid and should
	 *         be dropped from the pool, true otherwise.
	 */
	ValidateObject(ctx context.Context, object *PooledObject) bool

	/**
	 * Reinitialize an instance to be returned by the pool.
	 *
	 * return error if there is a problem activating object,
	 *    this error may be swallowed by the pool.
	 */
	ActivateObject(ctx context.Context, object *PooledObject) error

	/**
	 * Uninitialize an instance to be returned to the idle object pool.
	 *
	 * return error if there is a problem passivating obj,
	 *    this exception may be swallowed by the pool.
	 */
	PassivateObject(ctx context.Context, object *PooledObject) error
}

// DefaultPooledObjectFactory is a default PooledObjectFactory impl, support init by func
type DefaultPooledObjectFactory struct {
	make      func(ctx context.Context) (*PooledObject, error)
	destroy   func(ctx context.Context, object *PooledObject) error
	validate  func(ctx context.Context, object *PooledObject) bool
	activate  func(ctx context.Context, object *PooledObject) error
	passivate func(ctx context.Context, object *PooledObject) error
}

// NewPooledObjectFactorySimple return a DefaultPooledObjectFactory, only custom MakeObject func
func NewPooledObjectFactorySimple(
	create func(context.Context) (interface{}, error)) PooledObjectFactory {
	return NewPooledObjectFactory(create, nil, nil, nil, nil)
}

// NewPooledObjectFactory return a DefaultPooledObjectFactory, init with gaven func.
func NewPooledObjectFactory(
	create func(context.Context) (interface{}, error),
	destroy func(ctx context.Context, object *PooledObject) error,
	validate func(ctx context.Context, object *PooledObject) bool,
	activate func(ctx context.Context, object *PooledObject) error,
	passivate func(ctx context.Context, object *PooledObject) error) PooledObjectFactory {
	if create == nil {
		panic(errors.New("make function can not be nil"))
	}
	return &DefaultPooledObjectFactory{
		make: func(ctx context.Context) (*PooledObject, error) {
			o, err := create(ctx)
			if err != nil {
				return nil, err
			}
			return NewPooledObject(o), nil
		},
		destroy:   destroy,
		validate:  validate,
		activate:  activate,
		passivate: passivate}
}

// MakeObject see PooledObjectFactory.MakeObject
func (f *DefaultPooledObjectFactory) MakeObject(ctx context.Context) (*PooledObject, error) {
	return f.make(ctx)
}

// DestroyObject see PooledObjectFactory.DestroyObject
func (f *DefaultPooledObjectFactory) DestroyObject(ctx context.Context, object *PooledObject) error {
	if f.destroy != nil {
		return f.destroy(ctx, object)
	}
	return nil
}

// ValidateObject see PooledObjectFactory.ValidateObject
func (f *DefaultPooledObjectFactory) ValidateObject(ctx context.Context, object *PooledObject) bool {
	if f.validate != nil {
		return f.validate(ctx, object)
	}
	return true
}

// ActivateObject see PooledObjectFactory.ActivateObject
func (f *DefaultPooledObjectFactory) ActivateObject(ctx context.Context, object *PooledObject) error {
	if f.activate != nil {
		return f.activate(ctx, object)
	}
	return nil
}

// PassivateObject see PooledObjectFactory.PassivateObject
func (f *DefaultPooledObjectFactory) PassivateObject(ctx context.Context, object *PooledObject) error {
	if f.passivate != nil {
		return f.passivate(ctx, object)
	}
	return nil
}
