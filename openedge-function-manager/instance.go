package main

import (
	"fmt"
	"os"

	"github.com/baidu/openedge/sdk-go/openedge"
)

// Instance function instance
type Instance struct {
	n string
	f *Function
	c *openedge.FClient
}

// NewInstance creates a new instance
func (f *Function) NewInstance() (*Instance, error) {
	name, err := f.getID()
	if err != nil {
		f.log.WithError(err).Errorf("failed to create new instance")
		return nil, err
	}
	host, port := fmt.Sprintf("%s.%s", f.cfg.Service, name), "50051"
	if os.Getenv(openedge.EnvRunningModeKey) != "docker" {
		var err error
		host = "127.0.0.1"
		port, err = f.ctx.GetAvailablePort()
		if err != nil {
			return nil, err
		}
	}

	dc := make(map[string]string)
	dc[openedge.EnvServiceNameKey] = host
	dc[openedge.EnvServiceAddressKey] = fmt.Sprintf("0.0.0.0:%s", port)
	err = f.ctx.StartServiceInstance(f.cfg.Service, name, dc)
	if err != nil {
		return nil, err
	}
	fcc := openedge.FunctionClientConfig{}
	fcc.Address = fmt.Sprintf("%s:%s", host, port)
	fcc.Message = f.cfg.Message
	fcc.Timeout = f.cfg.Timeout
	fcc.Backoff = f.cfg.Backoff
	cli, err := openedge.NewFClient(fcc)
	if err != nil {
		f.ctx.StopServiceInstance(f.cfg.Service, name)
		return nil, err
	}
	return &Instance{
		n: name,
		f: f,
		c: cli,
	}, nil
}

// Call calls instance
func (i *Instance) Call(msg *openedge.FunctionMessage) (*openedge.FunctionMessage, error) {
	return i.c.Call(msg)
}

// Close closes instance
func (i *Instance) Close() error {
	i.f.returnID(i.n)
	if i.c != nil {
		i.c.Close()
	}
	return i.f.ctx.StopServiceInstance(i.f.cfg.Service, i.n)
}
