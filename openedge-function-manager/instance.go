package main

import (
	"fmt"
	"io"
	"os"

	"github.com/baidu/openedge/sdk-go/openedge"
)

// Instance function instance interface
type Instance interface {
	Name() string
	Call(msg *openedge.FunctionMessage) (*openedge.FunctionMessage, error)
	io.Closer
}

// Producer function instance producer interface
type Producer interface {
	StartInstance(name string) (Instance, error)
	StopInstance(i Instance) error
}

type producer struct {
	ctx openedge.Context
	cfg FunctionInfo
}

func newProducer(ctx openedge.Context, cfg FunctionInfo) Producer {
	return &producer{ctx: ctx, cfg: cfg}
}

// StartInstance starts instance
func (p *producer) StartInstance(name string) (Instance, error) {
	fullName := fmt.Sprintf("%s.%s", p.cfg.Service, name)
	port := "50051"
	serverHost := "0.0.0.0"
	clientHost := fullName
	if os.Getenv(openedge.EnvRunningModeKey) == "native" {
		var err error
		port, err = p.ctx.GetAvailablePort()
		if err != nil {
			return nil, err
		}
		serverHost = "127.0.0.1"
		clientHost = serverHost
	}

	dc := make(map[string]string)
	dc[openedge.EnvServiceNameKey] = fullName
	dc[openedge.EnvServiceAddressKey] = fmt.Sprintf("%s:%s", serverHost, port)
	err := p.ctx.StartServiceInstance(p.cfg.Service, name, dc)
	if err != nil {
		return nil, err
	}
	fcc := openedge.FunctionClientConfig{}
	fcc.Address = fmt.Sprintf("%s:%s", clientHost, port)
	fcc.Message = p.cfg.Message
	fcc.Timeout = p.cfg.Timeout
	fcc.Backoff = p.cfg.Backoff
	cli, err := openedge.NewFClient(fcc)
	if err != nil {
		p.ctx.StopServiceInstance(p.cfg.Service, name)
		return nil, err
	}
	return &instance{
		name:    name,
		FClient: cli,
	}, nil
}

// StopInstance stops instance
func (p *producer) StopInstance(i Instance) error {
	i.Close()
	return p.ctx.StopServiceInstance(p.cfg.Service, i.Name())
}

type instance struct {
	name string
	*openedge.FClient
}

// Name returns name
func (i *instance) Name() string {
	return i.name
}
