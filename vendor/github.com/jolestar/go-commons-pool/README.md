Go Commons Pool
=====

[![Build Status](https://travis-ci.org/jolestar/go-commons-pool.svg?branch=master)](https://travis-ci.org/jolestar/go-commons-pool)
[![CodeCov](https://codecov.io/gh/jolestar/go-commons-pool/branch/master/graph/badge.svg)](https://codecov.io/gh/jolestar/go-commons-pool)
[![Go Report Card](https://goreportcard.com/badge/github.com/jolestar/go-commons-pool)](https://goreportcard.com/report/github.com/jolestar/go-commons-pool)
[![GoDoc](https://godoc.org/github.com/jolestar/go-commons-pool?status.svg)](https://godoc.org/github.com/jolestar/go-commons-pool)

The Go Commons Pool is a generic object pool for [Golang](https://golang.org/), direct rewrite from [Apache Commons Pool](https://commons.apache.org/proper/commons-pool/).

Features
-------

1. Support custom [PooledObjectFactory](https://godoc.org/github.com/jolestar/go-commons-pool#PooledObjectFactory).
1. Rich pool configuration option, can precise control pooled object lifecycle. See [ObjectPoolConfig](https://godoc.org/github.com/jolestar/go-commons-pool#ObjectPoolConfig).
    * Pool LIFO (last in, first out) or FIFO (first in, first out)
    * Pool cap config
    * Pool object validate config
    * Pool object borrow block and max waiting time config
    * Pool object eviction config
    * Pool object abandon config

Pool Configuration Option
-------

Configuration option table, more detail description see [ObjectPoolConfig](https://godoc.org/github.com/jolestar/go-commons-pool#ObjectPoolConfig)

| Option                        | Default        | Description  |
| ------------------------------|:--------------:| :------------|
| LIFO                          | true           |If pool is LIFO (last in, first out)|
| MaxTotal                      | 8              |The cap of pool|
| MaxIdle                       | 8              |Max "idle" instances in the pool |
| MinIdle                       | 0              |Min "idle" instances in the pool |
| TestOnCreate                  | false          |Validate when object is created|
| TestOnBorrow                  | false          |Validate when object is borrowed|
| TestOnReturn                  | false          |Validate when object is returned|
| TestWhileIdle                 | false          |Validate when object is idle, see TimeBetweenEvictionRuns |
| BlockWhenExhausted            | true           |Whether to block when the pool is exhausted  |
| MinEvictableIdleTime          | 30m            |Eviction configuration,see DefaultEvictionPolicy |
| SoftMinEvictableIdleTime      | math.MaxInt64  |Eviction configuration,see DefaultEvictionPolicy  |
| NumTestsPerEvictionRun        | 3              |The maximum number of objects to examine during each run evictor goroutine |
| TimeBetweenEvictionRuns       | 0              |The number of milliseconds to sleep between runs of the evictor goroutine, less than 1 mean not run |

Usage
-------

### Use Simple Factory

```go
import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/jolestar/go-commons-pool"
)

func Example_simple() {
	type myPoolObject struct {
		s string
	}

	v := uint64(0)
	factory := pool.NewPooledObjectFactorySimple(
		func(context.Context) (interface{}, error) {
			return &myPoolObject{
					s: strconv.FormatUint(atomic.AddUint64(&v, 1), 10),
				},
				nil
		})

	ctx := context.Background()
	p := pool.NewObjectPoolWithDefaultConfig(ctx, factory)

	obj, err := p.BorrowObject(ctx)
	if err != nil {
		panic(err)
	}

	o := obj.(*myPoolObject)
	fmt.Println(o.s)

	err = p.ReturnObject(ctx, obj)
	if err != nil {
		panic(err)
	}

	// Output: 1
}
```

### Use Custom Factory

```go
import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/jolestar/go-commons-pool"
)

type MyPoolObject struct {
	s string
}

type MyCustomFactory struct {
	v uint64
}

func (f *MyCustomFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	return pool.NewPooledObject(
			&MyPoolObject{
				s: strconv.FormatUint(atomic.AddUint64(&f.v, 1), 10),
			}),
		nil
}

func (f *MyCustomFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	// do destroy
	return nil
}

func (f *MyCustomFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	// do validate
	return true
}

func (f *MyCustomFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	// do activate
	return nil
}

func (f *MyCustomFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	// do passivate
	return nil
}

func Example_customFactory() {
	ctx := context.Background()
	p := pool.NewObjectPoolWithDefaultConfig(ctx, &MyCustomFactory{})
	p.Config.MaxTotal = 100
    
	obj1, err := p.BorrowObject(ctx)
	if err != nil {
		panic(err)
	}

	o := obj1.(*MyPoolObject)
	fmt.Println(o.s)

	err = p.ReturnObject(ctx, obj1)
	if err != nil {
		panic(err)
	}

	// Output: 1
}
```

For more examples please see `pool_test.go` and `example_simple_test.go`, `example_customFactory_test.go`.

Note
-------

PooledObjectFactory.MakeObject must return a pointer, not value.
The following code will complain error.

```golang
p := pool.NewObjectPoolWithDefaultConfig(ctx, pool.NewPooledObjectFactorySimple(
    func(context.Context) (interface{}, error) {
        return "hello", nil
    }))
obj, _ := p.BorrowObject()
p.ReturnObject(obj)
```

The right way is:

```golang
p := pool.NewObjectPoolWithDefaultConfig(ctx, pool.NewPooledObjectFactorySimple(
    func(context.Context) (interface{}, error) {
        s := "hello"
        return &s, nil
    }))
```

For more examples please see `example_simple_test.go`.

Dependency
-------

* [testify](https://github.com/stretchr/testify) for test

PerformanceTest
-------

The results of running the pool_perf_test is almost equal to the java version [PerformanceTest](https://github.com/apache/commons-pool/blob/trunk/src/test/java/org/apache/commons/pool2/performance/PerformanceTest.java)

    go test --perf=true

For Apache commons pool user
-------

* Direct use pool.Config.xxx to change pool config
* Default config value is same as java version
* If TimeBetweenEvictionRuns changed after ObjectPool created, should call  **ObjectPool.StartEvictor** to take effect. Java version do this on set method.
* No KeyedObjectPool (TODO)
* No ProxiedObjectPool
* No pool stats (TODO)

FAQ
-------

[FAQ](https://github.com/jolestar/go-commons-pool/wiki/FAQ)

How to contribute
-------

* Choose one open issue you want to solve, if not create one and describe what you want to change.
* Fork the repository on GitHub.
* Write code to solve the issue.
* Create PR and link to the issue.
* Make sure test and coverage pass.
* Wait maintainers to merge.

中文文档
-------

* [Go Commons Pool发布以及Golang多线程编程问题总结](http://jolestar.com/go-commons-pool-and-go-concurrent/)
* [Go Commons Pool 1.0 发布](http://jolestar.com/go-commons-pool-v1-release/)

License
-------

Go Commons Pool is available under the [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0.html).
