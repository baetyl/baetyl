package pool

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"
)

const (
	// DefaultMaxTotal is the default value of ObjectPoolConfig.MaxTotal
	DefaultMaxTotal = 8
	// DefaultMaxIdle is the default value of ObjectPoolConfig.MaxIdle
	DefaultMaxIdle = 8
	// DefaultMinIdle is the default value of ObjectPoolConfig.MinIdle
	DefaultMinIdle = 0
	// DefaultLIFO is the default value of ObjectPoolConfig.LIFO
	DefaultLIFO = true

	// TODO
	// DEFAULT_FAIRNESS = false

	// DefaultMinEvictableIdleTime is the default value of ObjectPoolConfig.MinEvictableIdleTime
	DefaultMinEvictableIdleTime = 30 * time.Minute
	// DefaultSoftMinEvictableIdleTime is the default value of ObjectPoolConfig.SoftMinEvictableIdleTime
	DefaultSoftMinEvictableIdleTime = time.Duration(math.MaxInt64)
	// DefaultNumTestsPerEvictionRun is the default value of ObjectPoolConfig.NumTestsPerEvictionRun
	DefaultNumTestsPerEvictionRun = 3
	// DefaultTestOnCreate is the default value of ObjectPoolConfig.TestOnCreate
	DefaultTestOnCreate = false
	// DefaultTestOnBorrow is the default value of ObjectPoolConfig.TestOnBorrow
	DefaultTestOnBorrow = false
	// DefaultTestOnReturn is the default value of ObjectPoolConfig.TestOnReturn
	DefaultTestOnReturn = false
	// DefaultTestWhileIdle is the default value of ObjectPoolConfig.TestWhileIdle
	DefaultTestWhileIdle = false
	// DefaultTimeBetweenEvictionRuns is the default value of ObjectPoolConfig.TimeBetweenEvictionRuns
	DefaultTimeBetweenEvictionRuns = time.Duration(0)
	// DefaultBlockWhenExhausted is the default value of ObjectPoolConfig.BlockWhenExhausted
	DefaultBlockWhenExhausted = true
	// DefaultEvictionPolicyName is the default value of ObjectPoolConfig.EvictionPolicyName
	DefaultEvictionPolicyName = "github.com/jolestar/go-commons-pool/DefaultEvictionPolicy"
)

// ObjectPoolConfig is ObjectPool config, include cap, block, valid strategy, evict strategy etc.
type ObjectPoolConfig struct {
	/**
	 * Whether the pool has LIFO (last in, first out) behaviour with
	 * respect to idle objects - always returning the most recently used object
	 * from the pool, or as a FIFO (first in, first out) queue, where the pool
	 * always returns the oldest object in the idle object pool.
	 */
	LIFO bool

	/**
	 * The cap on the number of objects that can be allocated by the pool
	 * (checked out to clients, or idle awaiting checkout) at a given time. Use
	 * a negative value for no limit.
	 */
	MaxTotal int

	/**
	 * The cap on the number of "idle" instances in the pool. Use a
	 * negative value to indicate an unlimited number of idle instances.
	 * If MaxIdle
	 * is set too low on heavily loaded systems it is possible you will see
	 * objects being destroyed and almost immediately new objects being created.
	 * This is a result of the active goroutines momentarily returning objects
	 * faster than they are requesting them them, causing the number of idle
	 * objects to rise above maxIdle. The best value for maxIdle for heavily
	 * loaded system will vary but the default is a good starting point.
	 */
	MaxIdle int

	/**
	 * The minimum number of idle objects to maintain in
	 * the pool. This setting only has an effect if
	 * TimeBetweenEvictionRuns is greater than zero. If this
	 * is the case, an attempt is made to ensure that the pool has the required
	 * minimum number of instances during idle object eviction runs.
	 *
	 * If the configured value of MinIdle is greater than the configured value
	 * for MaxIdle then the value of MaxIdle will be used instead.
	 *
	 */
	MinIdle int

	/**
	* Whether objects created for the pool will be validated before
	* being returned from the ObjectPool.BorrowObject() method. Validation is
	* performed by the ValidateObject() method of the factory
	* associated with the pool. If the object fails to validate, then
	* ObjectPool.BorrowObject() will fail.
	 */
	TestOnCreate bool

	/**
	 * Whether objects borrowed from the pool will be validated before
	 * being returned from the ObjectPool.BorrowObject() method. Validation is
	 * performed by the ValidateObject() method of the factory
	 * associated with the pool. If the object fails to validate, it will be
	 * removed from the pool and destroyed, and a new attempt will be made to
	 * borrow an object from the pool.
	 */
	TestOnBorrow bool

	/**
	 * Whether objects borrowed from the pool will be validated when
	 * they are returned to the pool via the ObjectPool.ReturnObject() method.
	 * Validation is performed by the ValidateObject() method of
	 * the factory associated with the pool. Returning objects that fail validation
	 * are destroyed rather then being returned the pool.
	 */
	TestOnReturn bool

	/**
	* Whether objects sitting idle in the pool will be validated by the
	* idle object evictor (if any - see
	*  TimeBetweenEvictionRuns ). Validation is performed
	* by the ValidateObject() method of the factory associated
	* with the pool. If the object fails to validate, it will be removed from
	* the pool and destroyed.  Note that setting this property has no effect
	* unless the idle object evictor is enabled by setting
	* TimeBetweenEvictionRuns to a positive value.
	 */
	TestWhileIdle bool

	/**
	* Whether to block when the ObjectPool.BorrowObject() method is
	* invoked when the pool is exhausted (the maximum number of "active"
	* objects has been reached).
	 */
	BlockWhenExhausted bool

	//TODO support fairness config
	//Fairness                       bool

	/**
	 * The minimum amount of time an object may sit idle in the pool
	 * before it is eligible for eviction by the idle object evictor (if any -
	 * see TimeBetweenEvictionRuns. When non-positive,
	 * no objects will be evicted from the pool due to idle time alone.
	 */
	MinEvictableIdleTime time.Duration

	/**
	 * The minimum amount of time an object may sit idle in the pool
	 * before it is eligible for eviction by the idle object evictor (if any -
	 * see TimeBetweenEvictionRuns),
	 * with the extra condition that at least MinIdle object
	 * instances remain in the pool. This setting is overridden by
	 *  MinEvictableIdleTime (that is, if
	 *  MinEvictableIdleTime is positive, then
	 *  SoftMinEvictableIdleTime is ignored).
	 */
	SoftMinEvictableIdleTime time.Duration

	/**
	 * The maximum number of objects to examine during each run (if any)
	 * of the idle object evictor goroutine. When positive, the number of tests
	 * performed for a run will be the minimum of the configured value and the
	 * number of idle instances in the pool. When negative, the number of tests
	 * performed will be math.Ceil(ObjectPool.GetNumIdle()/math.
	 * Abs(PoolConfig.NumTestsPerEvictionRun)) which means that when the
	 * value is -n roughly one nth of the idle objects will be
	 * tested per run.
	 */
	NumTestsPerEvictionRun int

	/**
	 * The name of the EvictionPolicy implementation that is
	 * used by this pool. Please register policy by RegistryEvictionPolicy(name, policy)
	 */
	EvictionPolicyName string

	/**
	* The amount of time sleep between runs of the idle
	* object evictor goroutine. When non-positive, no idle object evictor goroutine
	* will be run.
	* if this value changed after ObjectPool created, should call ObjectPool.StartEvictor to take effect.
	 */
	TimeBetweenEvictionRuns time.Duration

	/**
	 * The context.Context to use when the evictor runs in the background.
	 */
	EvitionContext context.Context
}

// NewDefaultPoolConfig return a ObjectPoolConfig instance init with default value.
func NewDefaultPoolConfig() *ObjectPoolConfig {
	return &ObjectPoolConfig{
		LIFO:                     DefaultLIFO,
		MaxTotal:                 DefaultMaxTotal,
		MaxIdle:                  DefaultMaxIdle,
		MinIdle:                  DefaultMinIdle,
		MinEvictableIdleTime:     DefaultMinEvictableIdleTime,
		SoftMinEvictableIdleTime: DefaultSoftMinEvictableIdleTime,
		NumTestsPerEvictionRun:   DefaultNumTestsPerEvictionRun,
		EvictionPolicyName:       DefaultEvictionPolicyName,
		EvitionContext:           context.Background(),
		TestOnCreate:             DefaultTestOnCreate,
		TestOnBorrow:             DefaultTestOnBorrow,
		TestOnReturn:             DefaultTestOnReturn,
		TestWhileIdle:            DefaultTestWhileIdle,
		TimeBetweenEvictionRuns:  DefaultTimeBetweenEvictionRuns,
		BlockWhenExhausted:       DefaultBlockWhenExhausted}
}

// AbandonedConfig ObjectPool abandoned strategy config
type AbandonedConfig struct {
	RemoveAbandonedOnBorrow      bool
	RemoveAbandonedOnMaintenance bool
	// Timeout before an abandoned object can be removed.
	RemoveAbandonedTimeout time.Duration
}

// NewDefaultAbandonedConfig return a new AbandonedConfig instance init with default.
func NewDefaultAbandonedConfig() *AbandonedConfig {
	return &AbandonedConfig{RemoveAbandonedOnBorrow: false, RemoveAbandonedOnMaintenance: false, RemoveAbandonedTimeout: 5 * time.Minute}
}

// EvictionConfig is config for ObjectPool EvictionPolicy
type EvictionConfig struct {
	IdleEvictTime     time.Duration
	IdleSoftEvictTime time.Duration
	MinIdle           int
	Context           context.Context
}

// EvictionPolicy is a interface support custom EvictionPolicy
type EvictionPolicy interface {
	// Evict do evict by config
	Evict(config *EvictionConfig, underTest *PooledObject, idleCount int) bool
}

// DefaultEvictionPolicy is a default EvictionPolicy impl
type DefaultEvictionPolicy struct {
}

// Evict do evict by config
func (p *DefaultEvictionPolicy) Evict(config *EvictionConfig, underTest *PooledObject, idleCount int) bool {
	idleTime := underTest.GetIdleTime()

	if (config.IdleSoftEvictTime < idleTime &&
		config.MinIdle < idleCount) ||
		config.IdleEvictTime < idleTime {
		return true
	}
	return false
}

var (
	policiesMutex sync.Mutex
	policies      = make(map[string]EvictionPolicy)
)

// RegistryEvictionPolicy registry a custom EvictionPolicy with gaven name.
func RegistryEvictionPolicy(name string, policy EvictionPolicy) {
	if name == "" || policy == nil {
		panic(errors.New("invalid argument"))
	}
	policiesMutex.Lock()
	policies[name] = policy
	policiesMutex.Unlock()
}

// GetEvictionPolicy return a EvictionPolicy by gaven name
func GetEvictionPolicy(name string) EvictionPolicy {
	policiesMutex.Lock()
	defer policiesMutex.Unlock()
	return policies[name]

}

func init() {
	RegistryEvictionPolicy(DefaultEvictionPolicyName, new(DefaultEvictionPolicy))
}
