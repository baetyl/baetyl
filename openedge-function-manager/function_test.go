package main

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/baidu/openedge/logger"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	"github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFunction_CallAsync(t *testing.T) {
	cfg := FunctionInfo{Name: "test", Service: "dummy"}
	err := utils.SetDefaults(&cfg)
	assert.NoError(t, err)
	f := NewFunction(cfg, &mockProducer{})
	assert.NotNil(t, f)
	assert.Equal(t, 0, f.pool.GetNumActive())
	assert.Equal(t, 0, f.pool.GetNumIdle())
	assert.Equal(t, uint32(0), f.p.(*mockProducer).count)

	in := &openedge.FunctionMessage{
		ID:               1,
		QOS:              1,
		Topic:            "test",
		Payload:          []byte("a"),
		FunctionName:     "t",
		FunctionInvokeID: uuid.Generate().String(),
	}
	out, err := f.Call(in)
	assert.NoError(t, err)
	assert.Equal(t, in, out)
	assert.Equal(t, 0, f.pool.GetNumActive())
	assert.Equal(t, 1, f.pool.GetNumIdle())
	assert.Equal(t, uint32(1), f.p.(*mockProducer).count)

	in.Topic = "delay"
	finish := make(chan struct{})
	err = f.CallAsync(in, func(in, out *openedge.FunctionMessage, err error) {
		assert.NoError(t, err)
		assert.Equal(t, in, out)
		close(finish)
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, f.pool.GetNumActive())
	assert.Equal(t, 0, f.pool.GetNumIdle())
	<-finish
	assert.Equal(t, 0, f.pool.GetNumActive())
	assert.Equal(t, 1, f.pool.GetNumIdle())
	assert.Equal(t, 0, f.pool.GetDestroyedByBorrowValidationCount())
	assert.Equal(t, 0, f.pool.GetDestroyedCount())
	assert.Equal(t, uint32(1), f.p.(*mockProducer).count)

	in.Topic = "error"
	finish = make(chan struct{})
	err = f.CallAsync(in, func(in, out *openedge.FunctionMessage, err error) {
		assert.EqualError(t, err, "error")
		assert.Nil(t, out)
		close(finish)
	})
	assert.NoError(t, err)
	<-finish
	assert.Equal(t, 0, f.pool.GetNumActive())
	assert.Equal(t, 0, f.pool.GetNumIdle())
	assert.Equal(t, 0, f.pool.GetDestroyedByBorrowValidationCount())
	assert.Equal(t, 1, f.pool.GetDestroyedCount())
	assert.Equal(t, uint32(1), f.p.(*mockProducer).count)

	f.Close()
	_, err = f.Call(in)
	assert.EqualError(t, err, "Pool not open")
	err = f.CallAsync(in, nil)
	assert.EqualError(t, err, "Pool not open")
}

func TestFunctionPerf_Call(t *testing.T) {
	cfg := FunctionInfo{Name: "test", Service: "dummy"}
	err := utils.SetDefaults(&cfg)
	cfg.Instance.Max = 100
	assert.NoError(t, err)
	f := NewFunction(cfg, &mockProducer{})
	assert.NotNil(t, f)
	defer f.Close()
	assert.Equal(t, 0, f.pool.GetNumActive())
	assert.Equal(t, 0, f.pool.GetNumIdle())
	assert.Equal(t, uint32(0), f.p.(*mockProducer).count)

	in := &openedge.FunctionMessage{
		ID:               1,
		QOS:              1,
		Topic:            "test",
		Payload:          []byte("a"),
		FunctionName:     "t",
		FunctionInvokeID: uuid.Generate().String(),
	}
	for index := 0; index < 10; index++ {
		f.Call(in)
	}
	assert.Equal(t, 0, f.pool.GetNumActive())
	assert.Equal(t, 1, f.pool.GetNumIdle())
	assert.Equal(t, uint32(1), f.p.(*mockProducer).count)
}

func TestFunctionPerf_CallAsync(t *testing.T) {
	cfg := FunctionInfo{Name: "test", Service: "dummy"}
	err := utils.SetDefaults(&cfg)
	cfg.Instance.Max = 100
	assert.NoError(t, err)
	f := NewFunction(cfg, &mockProducer{})
	assert.NotNil(t, f)
	defer f.Close()
	assert.Equal(t, 0, f.pool.GetNumActive())
	assert.Equal(t, 0, f.pool.GetNumIdle())

	queue := make(chan *openedge.FunctionMessage, 100)
	in := &openedge.FunctionMessage{
		ID:               1,
		QOS:              1,
		Topic:            "delay",
		Payload:          []byte("a"),
		FunctionName:     "t",
		FunctionInvokeID: uuid.Generate().String(),
	}
	cb := func(in, out *openedge.FunctionMessage, err error) {
		queue <- out
	}
	count := 1000
	for index := 0; index < count; index++ {
		f.CallAsync(in, cb)
	}
	for index := 0; index < count; index++ {
		<-queue
	}
	assert.Equal(t, 0, f.pool.GetNumActive())
	assert.Equal(t, 100, f.pool.GetNumIdle())
	assert.Equal(t, uint32(100), f.p.(*mockProducer).count)
}

type mockContext struct {
}

func (c *mockContext) Config() *openedge.ServiceConfig {
	return nil
}
func (c *mockContext) UpdateSystem(*openedge.AppConfig) error {
	return nil
}
func (c *mockContext) InspectSystem() (*openedge.Inspect, error) {
	return nil, nil
}
func (c *mockContext) Log() logger.Logger {
	return nil
}
func (c *mockContext) Wait() {
}

func (c *mockContext) GetAvailablePort() (string, error) {
	return "", nil
}
func (c *mockContext) StartServiceInstance(serviceName, instanceName string, dynamicConfig map[string]string) error {
	return nil
}
func (c *mockContext) StopServiceInstance(serviceName, instanceName string) error {
	return nil
}

type mockProducer struct {
	count uint32
}

func (p *mockProducer) StartInstance(name string) (Instance, error) {
	return &mockInstance{index: atomic.AddUint32(&p.count, 1)}, nil
}

func (p *mockProducer) StopInstance(i Instance) error {
	return nil
}

type mockInstance struct {
	index uint32
}

func (i *mockInstance) Name() string {
	return ""
}
func (i *mockInstance) Call(msg *openedge.FunctionMessage) (*openedge.FunctionMessage, error) {
	if msg.Topic == "delay" {
		time.Sleep(50 * time.Millisecond)
	} else if msg.Topic == "error" {
		return nil, fmt.Errorf("error")
	}
	return msg, nil
}

func (i *mockInstance) Close() error {
	return nil
}
