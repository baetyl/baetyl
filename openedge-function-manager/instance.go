package main

import (
	"fmt"
	"io"
	"os"

	openedge "github.com/baidu/openedge/sdk/openedge-go"
)

// Instance function instance interface
type Instance interface {
	ID() uint32
	Name() string
	Call(msg *openedge.FunctionMessage) (*openedge.FunctionMessage, error)
	io.Closer
}

// Producer function instance producer interface
type Producer interface {
	StartInstance(id uint32) (Instance, error)
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
func (p *producer) StartInstance(id uint32) (Instance, error) {
	name := fmt.Sprintf("%s.%s.%d", p.cfg.Service, p.cfg.Name, id)
	port := "50051"
	serverHost := "0.0.0.0"
	clientHost := name
	if os.Getenv(openedge.EnvRunningModeKey) == "native" {
		var err error
		port, err = p.ctx.GetAvailablePort()
		if err != nil {
			return nil, err
		}
		serverHost = "127.0.0.1"
		clientHost = serverHost
	}

	address := fmt.Sprintf("%s:%s", serverHost, port)
	dc := map[string]string{
		openedge.EnvServiceAddressKey:         address, // deprecated, for v0.1.2
		openedge.EnvServiceInstanceAddressKey: address,
	}
	err := p.ctx.StartInstance(p.cfg.Service, name, dc)
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
		p.ctx.StopInstance(p.cfg.Service, name)
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
	return p.ctx.StopInstance(p.cfg.Service, i.Name())
}

type instance struct {
	id   uint32
	name string
	*openedge.FClient
}

// ID returns id
func (i *instance) ID() uint32 {
	return i.id
}

// Name returns name
func (i *instance) Name() string {
	return i.name
}
