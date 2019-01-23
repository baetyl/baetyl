package jrpc

import (
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/baidu/openedge/utils"
)

// Client client of http server
type Client struct {
	*rpc.Client
}

// NewClient creates a new http client
func NewClient(addr string) (*Client, error) {
	url, err := utils.ParseURL(addr)
	if err != nil {
		return nil, err
	}
	cli, err := jsonrpc.Dial(url.Scheme, url.Host)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client: cli,
	}, nil
}
