package openedge

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/baidu/openedge/protocol/http"
)

// Client client of api server
type Client struct {
	cli *http.Client
}

// NewEnvClient creates a new client by env
func NewEnvClient() (*Client, error) {
	addr := os.Getenv(EnvMasterAPIKey)
	if len(addr) == 0 {
		return nil, fmt.Errorf("Env (%s) not found", EnvMasterAPIKey)
	}
	c := http.ClientInfo{
		Address:  addr,
		Username: os.Getenv(EnvServiceNameKey),
		Password: os.Getenv(EnvServiceTokenKey),
	}
	return NewClient(c)
}

// NewClient creates a new client
func NewClient(c http.ClientInfo) (*Client, error) {
	cli, err := http.NewClient(c)
	if err != nil {
		return nil, err
	}
	return &Client{
		cli: cli,
	}, nil
}

// InspectSystem inspect all stats
func (c *Client) InspectSystem() (*Inspect, error) {
	body, err := c.cli.Get("/inspect")
	if err != nil {
		return nil, err
	}
	s := new(Inspect)
	err = json.Unmarshal(body, s)
	return s, err
}

// UpdateSystem updates and reloads config
func (c *Client) UpdateSystem(d *DatasetInfo) error {
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	_, err = c.cli.Put(data, "/update")
	return err
}
