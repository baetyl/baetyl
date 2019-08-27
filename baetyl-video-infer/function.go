package main

import (
	"fmt"

	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/docker/distribution/uuid"
)

// Functions function clients
type Functions struct {
	configs []FunctionInfo
	clients map[string]*baetyl.FClient
}

// NewFunctions creates a new Functions
func NewFunctions(configs []FunctionInfo) (*Functions, error) {
	clients := map[string]*baetyl.FClient{}
	for _, cfg := range configs {
		if cfg.Address == "" {
			continue
		}
		cli, err := baetyl.NewFClient(cfg.FunctionClientConfig)
		if err != nil {
			for _, item := range clients {
				item.Close()
			}
			return nil, fmt.Errorf("failed to create function client: %s", err.Error())
		}
		clients[cfg.Name] = cli
	}
	return &Functions{
		configs: configs,
		clients: clients,
	}, nil
}

// Call calls the function
func (f *Functions) Call(name string, ts int64, payload []byte) ([]byte, error) {
	cli, ok := f.clients[name]
	if !ok {
		return nil, fmt.Errorf("function (%s) not found", name)
	}
	in := &baetyl.FunctionMessage{
		Payload:          payload,
		Timestamp:        ts,
		FunctionName:     name,
		FunctionInvokeID: uuid.Generate().String(),
	}
	out, err := cli.Call(in)
	if err != nil {
		return nil, fmt.Errorf("[%s] %s", in.FunctionInvokeID, err.Error())
	}
	if out == nil {
		return nil, nil
	}
	return out.Payload, nil
}
