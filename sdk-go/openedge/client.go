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
	body, err := c.cli.Get("/system/inspect")
	if err != nil {
		return nil, err
	}
	s := new(Inspect)
	err = json.Unmarshal(body, s)
	return s, err
}

// UpdateSystem updates and reloads config
func (c *Client) UpdateSystem(data []byte) error {
	_, err := c.cli.Put(data, "/system/update")
	return err
}

// GetAvailablePort gets available port
func (c *Client) GetAvailablePort() (string, error) {
	res, err := c.cli.Get("/ports/available")
	if err != nil {
		return "", err
	}
	info := make(map[string]string)
	err = json.Unmarshal(res, &info)
	if err != nil {
		return "", err
	}
	port, ok := info["port"]
	if !ok {
		return "", fmt.Errorf("invalid response, port not found")
	}
	return port, nil
}

// StartServiceInstance starts a new service instance with dynamic config
func (c *Client) StartServiceInstance(serviceName, instanceName string, dynamicConfig map[string]string) error {
	data, err := json.Marshal(dynamicConfig)
	if err != nil {
		return err
	}
	_, err = c.cli.Put(data, "/services/%s/instances/%s/start", serviceName, instanceName)
	return err
}

// StopServiceInstance stops a service instance
func (c *Client) StopServiceInstance(serviceName, instanceName string) error {
	_, err := c.cli.Put(nil, "/services/%s/instances/%s/stop", serviceName, instanceName)
	return err
}
