package main

import (
	"fmt"
	"io"
	"os"

	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

// Instance function instance interface
type Instance interface {
	ID() uint32
	Name() string
	Call(msg *baetyl.FunctionMessage) (*baetyl.FunctionMessage, error)
	io.Closer
}

// Producer function instance producer interface
type Producer interface {
	StartInstance(id uint32) (Instance, error)
	StopInstance(i Instance) error
}

type producer struct {
	ctx baetyl.Context
	cfg FunctionInfo
}

func newProducer(ctx baetyl.Context, cfg FunctionInfo) Producer {
	return &producer{ctx: ctx, cfg: cfg}
}

// StartInstance starts instance
func (p *producer) StartInstance(id uint32) (Instance, error) {
	name := fmt.Sprintf("%s.%s.%d", p.cfg.Service, p.cfg.Name, id)
	port := "50051"
	serverHost := "0.0.0.0"
	clientHost := name
	if os.Getenv(baetyl.EnvKeyServiceMode) == "native" ||
		/*backward compatibility*/ os.Getenv(baetyl.EnvRunningModeKey) == "native" {
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
		baetyl.EnvServiceAddressKey:         address, // deprecated, for v0.1.2
		baetyl.EnvServiceInstanceAddressKey: address, // deprecated, for v0.1.3~5
		baetyl.EnvKeyServiceInstanceAddress: address,
	}
	err := p.ctx.StartInstance(p.cfg.Service, name, dc)
	if err != nil {
		return nil, err
	}
	fcc := baetyl.FunctionClientConfig{}
	fcc.Address = fmt.Sprintf("%s:%s", clientHost, port)
	fcc.Message = p.cfg.Message
	fcc.Timeout = p.cfg.Timeout
	fcc.Backoff = p.cfg.Backoff
	cli, err := baetyl.NewFClient(fcc)
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
	*baetyl.FClient
}

// ID returns id
func (i *instance) ID() uint32 {
	return i.id
}

// Name returns name
func (i *instance) Name() string {
	return i.name
}
