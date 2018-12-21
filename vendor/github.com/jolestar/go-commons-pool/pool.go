package pool

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/jolestar/go-commons-pool/collections"
	"github.com/jolestar/go-commons-pool/concurrent"
)

type baseErr struct {
	msg string
}

func (err *baseErr) Error() string {
	return err.msg
}

// IllegalStateErr when use pool in a illegal way, return this err
type IllegalStateErr struct {
	baseErr
}

// NewIllegalStateErr return new IllegalStateErr
func NewIllegalStateErr(msg string) *IllegalStateErr {
	return &IllegalStateErr{baseErr{msg}}
}

// NoSuchElementErr when no available object in pool, return this err
type NoSuchElementErr struct {
	baseErr
}

// NewNoSuchElementErr return new NoSuchElementErr
func NewNoSuchElementErr(msg string) *NoSuchElementErr {
	return &NoSuchElementErr{baseErr{msg}}
}

// ObjectPool is a generic object pool
type ObjectPool struct {
	AbandonedConfig                  *AbandonedConfig
	Config                           *ObjectPoolConfig
	closed                           bool
	closeLock                        sync.Mutex
	evictionLock                     sync.Mutex
	idleObjects                      *collections.LinkedBlockingDeque
	allObjects                       *collections.SyncIdentityMap
	factory                          PooledObjectFactory
	createCount                      concurrent.AtomicInteger
	destroyedByEvictorCount          concurrent.AtomicInteger
	destroyedCount                   concurrent.AtomicInteger
	destroyedByBorrowValidationCount concurrent.AtomicInteger
	evictor                          *time.Ticker
	evictorStopChan                  chan struct{}
	evictorStopWG                    sync.WaitGroup
	evictionIterator                 collections.Iterator
}

// NewObjectPool return new ObjectPool, init with PooledObjectFactory and ObjectPoolConfig
func NewObjectPool(ctx context.Context, factory PooledObjectFactory, config *ObjectPoolConfig) *ObjectPool {
	return NewObjectPoolWithAbandonedConfig(ctx, factory, config, nil)
}

// NewObjectPoolWithDefaultConfig return new ObjectPool init with PooledObjectFactory and default config
func NewObjectPoolWithDefaultConfig(ctx context.Context, factory PooledObjectFactory) *ObjectPool {
	return NewObjectPool(ctx, factory, NewDefaultPoolConfig())
}

// NewObjectPoolWithAbandonedConfig return new ObjectPool init with PooledObjectFactory, ObjectPoolConfig, and AbandonedConfig
func NewObjectPoolWithAbandonedConfig(ctx context.Context, factory PooledObjectFactory, config *ObjectPoolConfig, abandonedConfig *AbandonedConfig) *ObjectPool {
	pool := ObjectPool{factory: factory, Config: config,
		idleObjects:             collections.NewDeque(math.MaxInt32),
		allObjects:              collections.NewSyncMap(),
		createCount:             concurrent.AtomicInteger(0),
		destroyedByEvictorCount: concurrent.AtomicInteger(0),
		destroyedCount:          concurrent.AtomicInteger(0),
		AbandonedConfig:         abandonedConfig}
	pool.StartEvictor()
	return &pool
}

// AddObject create an object using the PooledObjectFactory factory, passivate it, and then place it in
// the idle object pool. AddObject is useful for "pre-loading"
// a pool with idle objects. (Optional operation).
func (pool *ObjectPool) AddObject(ctx context.Context) error {
	if pool.IsClosed() {
		return NewIllegalStateErr("Pool not open")
	}
	if pool.factory == nil {
		return NewIllegalStateErr("Cannot add objects without a factory.")
	}
	p, e := pool.create(ctx)
	if e != nil {
		return e
	}
	e = pool.addIdleObject(ctx, p)
	if e != nil {
		pool.destroy(ctx, p)
		return e
	}
	return nil
}

func (pool *ObjectPool) addIdleObject(ctx context.Context, p *PooledObject) error {
	if p != nil {
		err := pool.factory.PassivateObject(ctx, p)
		if err != nil {
			return err
		}

		if pool.Config.LIFO {
			err = pool.idleObjects.AddFirst(p)
		} else {
			err = pool.idleObjects.AddLast(p)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// BorrowObject obtains an instance from pool.
// Instances returned from pool method will have been either newly created
// with PooledObjectFactory.MakeObject or will be a previously
// idle object and have been activated with
// PooledObjectFactory.ActivateObject and then validated with
// PooledObjectFactory.ValidateObject.
// If the pool is full (based on the number of objects in the pool and the
// value of the MaxTotal configuration field), this method will block until
// an object is returned to the pool or the context is done.
//
// By contract, clients must return the borrowed instance
// using ReturnObject, InvalidateObject
func (pool *ObjectPool) BorrowObject(ctx context.Context) (interface{}, error) {
	return pool.borrowObject(ctx)
}

// GetNumIdle return the number of instances currently idle in pool. This may be
// considered an approximation of the number of objects that can be
// BorrowObject borrowed without creating any new instances.
func (pool *ObjectPool) GetNumIdle() int {
	return pool.idleObjects.Size()
}

// GetNumActive return the number of instances currently borrowed from pool.
func (pool *ObjectPool) GetNumActive() int {
	return pool.allObjects.Size() - pool.idleObjects.Size()
}

// GetDestroyedCount return destroyed object count of this pool
func (pool *ObjectPool) GetDestroyedCount() int {
	return int(pool.destroyedCount.Get())
}

// GetDestroyedByBorrowValidationCount return destroyed object count when borrow validation
func (pool *ObjectPool) GetDestroyedByBorrowValidationCount() int {
	return int(pool.destroyedByBorrowValidationCount.Get())
}

func (pool *ObjectPool) removeAbandoned(ctx context.Context, config *AbandonedConfig) {
	// Generate a list of abandoned objects to remove
	var remove []*PooledObject
	objects := pool.allObjects.Values()
	for _, o := range objects {
		pooledObject := o.(*PooledObject)
		pooledObject.lock.Lock()
		if pooledObject.state == StateAllocated &&
			time.Since(pooledObject.GetLastUsedTime()) > config.RemoveAbandonedTimeout {
			pooledObject.markAbandoned()
			remove = append(remove, pooledObject)
		}
		pooledObject.lock.Unlock()
	}

	// Now remove the abandoned objects
	for _, pooledObject := range remove {
		//if (config.getLogAbandoned()) {
		//pooledObject.printStackTrace(ac.getLogWriter());
		//}
		pool.InvalidateObject(ctx, pooledObject.Object)
	}
}

func (pool *ObjectPool) create(ctx context.Context) (*PooledObject, error) {
	localMaxTotal := pool.Config.MaxTotal
	newCreateCount := pool.createCount.IncrementAndGet()
	if localMaxTotal > -1 && int(newCreateCount) > localMaxTotal ||
		newCreateCount >= math.MaxInt32 {
		pool.createCount.DecrementAndGet()
		return nil, nil
	}

	p, e := pool.factory.MakeObject(ctx)
	if e != nil {
		pool.createCount.DecrementAndGet()
		return nil, e
	}

	//	ac := pool.abandonedConfig;
	//	if (ac != null && ac.getLogAbandoned()) {
	//		p.setLogAbandoned(true);
	//	}
	pool.allObjects.Put(p.Object, p)
	return p, nil
}

func (pool *ObjectPool) destroy(ctx context.Context, toDestroy *PooledObject) {
	pool.doDestroy(ctx, toDestroy, false)
}

func (pool *ObjectPool) doDestroy(ctx context.Context, toDestroy *PooledObject, inLock bool) {
	//golang has not recursive lock, so ...
	if inLock {
		toDestroy.invalidate()
	} else {
		toDestroy.Invalidate()
	}
	pool.idleObjects.RemoveFirstOccurrence(toDestroy)
	pool.allObjects.Remove(toDestroy.Object)
	pool.factory.DestroyObject(ctx, toDestroy)
	pool.destroyedCount.IncrementAndGet()
	pool.createCount.DecrementAndGet()
}

func (pool *ObjectPool) updateStatsBorrow(object *PooledObject, time time.Duration) {
	//TODO
}

func (pool *ObjectPool) updateStatsReturn(activeTime time.Duration) {
	//TODO
	//returnedCount.incrementAndGet();
	//activeTimes.add(activeTime);
}

func (pool *ObjectPool) borrowObject(ctx context.Context) (interface{}, error) {
	if pool.IsClosed() {
		return nil, NewIllegalStateErr("Pool not open")
	}
	ac := pool.AbandonedConfig
	if ac != nil && ac.RemoveAbandonedOnBorrow &&
		(pool.GetNumIdle() < 2) &&
		(pool.GetNumActive() > pool.Config.MaxTotal-3) {
		pool.removeAbandoned(ctx, ac)
	}

	var p *PooledObject
	var e error
	// Get local copy of current config so it is consistent for entire
	// method execution
	blockWhenExhausted := pool.Config.BlockWhenExhausted

	var create bool
	waitTime := time.Now()
	var ok bool
	for p == nil {
		create = false
		if blockWhenExhausted {
			p, ok = pool.idleObjects.PollFirst().(*PooledObject)
			if !ok {
				p, e = pool.create(ctx)
				if e != nil {
					return nil, e
				}
				if p != nil {
					create = true
					ok = true
				}
			}
			if p == nil {
				obj, err := pool.idleObjects.PollFirstWithContext(ctx)
				if err != nil {
					return nil, err
				}
				p, ok = obj.(*PooledObject)
			}
			if !ok {
				return nil, NewNoSuchElementErr("Timeout waiting for idle object")
			}
			if !p.Allocate() {
				p = nil
			}
		} else {
			p, ok = pool.idleObjects.PollFirst().(*PooledObject)
			if !ok {
				p, e = pool.create(ctx)
				if e != nil {
					return nil, e
				}
				if p != nil {
					create = true
				}
			}
			if p == nil {
				return nil, NewNoSuchElementErr("Pool exhausted")
			}
			if !p.Allocate() {
				p = nil
			}
		}

		if p != nil {
			e := pool.factory.ActivateObject(ctx, p)
			if e != nil {
				pool.destroy(ctx, p)
				p = nil
				if create {
					return nil, NewNoSuchElementErr("Unable to activate object")
				}
			}
		}
		if p != nil && (pool.Config.TestOnBorrow || create && pool.Config.TestOnCreate) {
			validate := pool.factory.ValidateObject(ctx, p)
			if !validate {
				pool.destroy(ctx, p)
				pool.destroyedByBorrowValidationCount.IncrementAndGet()
				p = nil
				if create {
					return nil, NewNoSuchElementErr("Unable to validate object")
				}
			}
		}
	}

	pool.updateStatsBorrow(p, time.Since(waitTime))
	return p.Object, nil
}

func (pool *ObjectPool) isAbandonedConfig() bool {
	return pool.AbandonedConfig != nil
}

func (pool *ObjectPool) ensureIdle(ctx context.Context, idleCount int, always bool) {
	if idleCount < 1 || pool.IsClosed() || (!always && !pool.idleObjects.HasTakeWaiters()) {
		return
	}

	for pool.idleObjects.Size() < idleCount {
		//just ignore create error
		p, _ := pool.create(ctx)
		if p == nil {
			// Can't create objects, no reason to think another call to
			// create will work. Give up.
			break
		}
		if pool.Config.LIFO {
			pool.idleObjects.AddFirst(p)
		} else {
			pool.idleObjects.AddLast(p)
		}
	}
	if pool.IsClosed() {
		// Pool closed while object was being added to idle objects.
		// Make sure the returned object is destroyed rather than left
		// in the idle object pool (which would effectively be a leak)
		pool.Clear(ctx)
	}
}

// IsClosed return this pool is closed
func (pool *ObjectPool) IsClosed() bool {
	pool.closeLock.Lock()
	defer pool.closeLock.Unlock()
	// in java commons pool, closed is volatile, golang has not volatile, so use mutex to avoid data race
	return pool.closed
}

// ReturnObject return an instance to the pool. By contract, object
// must have been obtained using BorrowObject()
func (pool *ObjectPool) ReturnObject(ctx context.Context, object interface{}) error {
	if object == nil {
		return errors.New("object is nil")
	}
	p, ok := pool.allObjects.Get(object).(*PooledObject)

	if !ok {
		if !pool.isAbandonedConfig() {
			return NewIllegalStateErr(
				"Returned object not currently part of pool")
		}
		return nil // Object was abandoned and removed
	}
	p.lock.Lock()

	state := p.state
	if state != StateAllocated {
		p.lock.Unlock()
		return NewIllegalStateErr(
			"Object has already been returned to pool or is invalid")
	}
	//use unlock method markReturning() not MarkReturning
	// because go lock is not recursive
	p.markReturning() // Keep from being marked abandoned
	p.lock.Unlock()
	activeTime := p.GetActiveTime()

	if pool.Config.TestOnReturn {
		if !pool.factory.ValidateObject(ctx, p) {
			pool.destroy(ctx, p)
			pool.ensureIdle(ctx, 1, false)
			pool.updateStatsReturn(activeTime)
			// swallowException(e);
			return nil
		}
	}

	err := pool.factory.PassivateObject(ctx, p)
	if err != nil {
		//swallowException(e1);
		pool.destroy(ctx, p)
		pool.ensureIdle(ctx, 1, false)
		pool.updateStatsReturn(activeTime)
		// swallowException(e);
		return nil
	}

	if !p.Deallocate() {
		return NewIllegalStateErr("Object has already been returned to pool or is invalid")
	}

	maxIdleSave := pool.Config.MaxIdle
	if pool.IsClosed() || maxIdleSave > -1 && maxIdleSave <= pool.idleObjects.Size() {
		pool.destroy(ctx, p)
	} else {
		if pool.Config.LIFO {
			pool.idleObjects.AddFirst(p)
		} else {
			pool.idleObjects.AddLast(p)
		}
		if pool.IsClosed() {
			// Pool closed while object was being added to idle objects.
			// Make sure the returned object is destroyed rather than left
			// in the idle object pool (which would effectively be a leak)
			pool.Clear(ctx)
		}
	}
	pool.updateStatsReturn(activeTime)
	return nil
}

// Clear clears any objects sitting idle in the pool, releasing any associated
// resources (optional operation). Idle objects cleared must be
// PooledObjectFactory.DestroyObject(PooledObject) .
func (pool *ObjectPool) Clear(ctx context.Context) {
	p, ok := pool.idleObjects.PollFirst().(*PooledObject)

	for ok {
		pool.destroy(ctx, p)
		p, ok = pool.idleObjects.PollFirst().(*PooledObject)
	}
}

// InvalidateObject invalidates an object from the pool.
// By contract, object must have been obtained
// using BorrowObject.
// This method should be used when an object that has been borrowed is
// determined (due to an exception or other problem) to be invalid.
func (pool *ObjectPool) InvalidateObject(ctx context.Context, object interface{}) error {
	p, ok := pool.allObjects.Get(object).(*PooledObject)
	if !ok {
		if pool.isAbandonedConfig() {
			return nil
		}
		return NewIllegalStateErr(
			"Invalidated object not currently part of pool")
	}
	p.lock.Lock()
	if p.state != StateInvalid {
		pool.doDestroy(ctx, p, true)
	}
	p.lock.Unlock()
	pool.ensureIdle(ctx, 1, false)
	return nil
}

// Close pool, and free any resources associated with it.
func (pool *ObjectPool) Close(ctx context.Context) {
	if pool.IsClosed() {
		return
	}
	pool.closeLock.Lock()
	defer pool.closeLock.Unlock()
	if pool.closed {
		return
	}

	// Stop the evictor before the pool is closed since evict() calls
	// assertOpen()
	pool.startEvictor(-1)

	pool.closed = true
	// This clear removes any idle objects
	pool.Clear(ctx)

	// Release any goroutines that were waiting for an object
	pool.idleObjects.InterruptTakeWaiters()
}

// StartEvictor start background evictor goroutine, pool.Config.TimeBetweenEvictionRuns must a positive number.
// if ObjectPool.Config.TimeBetweenEvictionRuns change, should call pool method to let it to take effect.
func (pool *ObjectPool) StartEvictor() {
	pool.startEvictor(pool.Config.TimeBetweenEvictionRuns)
}

func (pool *ObjectPool) startEvictor(delay time.Duration) {
	pool.evictionLock.Lock()
	defer pool.evictionLock.Unlock()
	if pool.evictor != nil {
		pool.evictor.Stop()
		close(pool.evictorStopChan)
		//Ensure old evictor goroutine quit, only one evictor goroutine at same time, then set evictor to nil.
		pool.evictorStopWG.Wait()
		pool.evictor = nil
		pool.evictionIterator = nil
	}
	if delay > 0 {
		pool.evictor = time.NewTicker(delay)
		pool.evictorStopChan = make(chan struct{})
		pool.evictorStopWG = sync.WaitGroup{}
		pool.evictorStopWG.Add(1)
		go func() {
			for {
				select {
				case <-pool.evictor.C:
					pool.evict(pool.Config.EvitionContext)
					pool.ensureMinIdle(pool.Config.EvitionContext)
				case <-pool.evictorStopChan:
					pool.evictorStopWG.Done()
					return
				}
			}
		}()
	}
}

func (pool *ObjectPool) getEvictionPolicy() EvictionPolicy {
	evictionPolicy := GetEvictionPolicy(pool.Config.EvictionPolicyName)
	if evictionPolicy == nil {
		evictionPolicy = GetEvictionPolicy(DefaultEvictionPolicyName)
	}
	return evictionPolicy
}

func (pool *ObjectPool) getNumTests() int {
	numTestsPerEvictionRun := pool.Config.NumTestsPerEvictionRun
	if numTestsPerEvictionRun >= 0 {
		if numTestsPerEvictionRun < pool.idleObjects.Size() {
			return numTestsPerEvictionRun
		}
		return pool.idleObjects.Size()
	}
	return int((math.Ceil(float64(pool.idleObjects.Size()) / math.Abs(float64(numTestsPerEvictionRun)))))
}

// idleIterator return pool idleObjects iterator
func (pool *ObjectPool) idleIterator() collections.Iterator {
	if pool.Config.LIFO {
		return pool.idleObjects.DescendingIterator()
	}
	return pool.idleObjects.Iterator()
}

func (pool *ObjectPool) getMinIdle() int {
	maxIdleSave := pool.Config.MaxIdle
	if pool.Config.MinIdle > maxIdleSave {
		return maxIdleSave
	}
	return pool.Config.MinIdle
}

func (pool *ObjectPool) evict(ctx context.Context) {
	defer func() {
		ac := pool.AbandonedConfig
		if ac != nil && ac.RemoveAbandonedOnMaintenance {
			pool.removeAbandoned(ctx, ac)
		}
	}()

	if pool.idleObjects.Size() == 0 {
		return
	}
	var underTest *PooledObject
	evictionPolicy := pool.getEvictionPolicy()

	evictionConfig := EvictionConfig{
		IdleEvictTime:     pool.Config.MinEvictableIdleTime,
		IdleSoftEvictTime: pool.Config.SoftMinEvictableIdleTime,
		MinIdle:           pool.Config.MinIdle,
		Context:           pool.Config.EvitionContext}

	testWhileIdle := pool.Config.TestWhileIdle
	for i, m := 0, pool.getNumTests(); i < m; i++ {
		if pool.evictionIterator == nil || !pool.evictionIterator.HasNext() {
			pool.evictionIterator = pool.idleIterator()
		}
		if !pool.evictionIterator.HasNext() {
			// Pool exhausted, nothing to do here
			return
		}

		underTest = pool.evictionIterator.Next().(*PooledObject)
		if underTest == nil {
			// Object was borrowed in another goroutine
			// Don't count pool as an eviction test so reduce i;
			i--
			pool.evictionIterator = nil
			continue
		}

		if !underTest.StartEvictionTest() {
			// Object was borrowed in another goroutine
			// Don't count pool as an eviction test so reduce i;
			i--
			continue
		}

		// User provided eviction policy could throw all sorts of
		// crazy exceptions. Protect against such an exception
		// killing the eviction goroutine.

		evict := evictionPolicy.Evict(&evictionConfig, underTest, pool.idleObjects.Size())

		if evict {
			pool.destroy(ctx, underTest)
			pool.destroyedByEvictorCount.IncrementAndGet()
		} else {
			active := false
			if testWhileIdle {
				err := pool.factory.ActivateObject(ctx, underTest)
				if err == nil {
					active = true
				} else {
					pool.destroy(ctx, underTest)
					pool.destroyedByEvictorCount.IncrementAndGet()
				}
				if active {
					if !pool.factory.ValidateObject(ctx, underTest) {
						pool.destroy(ctx, underTest)
						pool.destroyedByEvictorCount.IncrementAndGet()
					} else {
						err := pool.factory.PassivateObject(ctx, underTest)
						if err != nil {
							pool.destroy(ctx, underTest)
							pool.destroyedByEvictorCount.IncrementAndGet()
						}
					}
				}
			}
			if !underTest.EndEvictionTest(pool.idleObjects) {
				// TODO - May need to add code here once additional
				// states are used
			}
		}
	}
}

func (pool *ObjectPool) ensureMinIdle(ctx context.Context) {
	pool.ensureIdle(ctx, pool.getMinIdle(), true)
}

// PreparePool Tries to ensure that {@link #getMinIdle()} idle instances are available
// in the pool.
func (pool *ObjectPool) PreparePool(ctx context.Context) {
	if pool.getMinIdle() < 1 {
		return
	}
	pool.ensureMinIdle(ctx)
}

// Prefill is util func for pre fill pool object
func Prefill(ctx context.Context, pool *ObjectPool, count int) {
	for i := 0; i < count; i++ {
		pool.AddObject(ctx)
	}
}
