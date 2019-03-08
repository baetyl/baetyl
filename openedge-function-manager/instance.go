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
	fullName := fmt.Sprintf("%s.%s", f.cfg.Service, name)
	port := "50051"
	serverHost := "0.0.0.0"
	clientHost := fullName
	if os.Getenv(openedge.EnvRunningModeKey) == "native" {
		var err error
		port, err = f.ctx.GetAvailablePort()
		if err != nil {
			return nil, err
		}
		serverHost = "127.0.0.1"
		clientHost = serverHost
	}

	dc := make(map[string]string)
	dc[openedge.EnvServiceNameKey] = fullName
	dc[openedge.EnvServiceAddressKey] = fmt.Sprintf("%s:%s", serverHost, port)
	err = f.ctx.StartServiceInstance(f.cfg.Service, name, dc)
	if err != nil {
		return nil, err
	}
	fcc := openedge.FunctionClientConfig{}
	fcc.Address = fmt.Sprintf("%s:%s", clientHost, port)
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
